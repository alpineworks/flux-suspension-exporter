package flux

import (
	"context"
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesClient struct {
	client client.Client
}

func NewKubernetesClient() (*KubernetesClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in cluster config: %w", err)
	}

	client, err := client.NewWithWatch(config, client.Options{
		Scheme: GetScheme(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &KubernetesClient{
		client: client,
	}, nil
}

func (kc *KubernetesClient) GetSuspendedKustomizations(ctx context.Context) ([]kustomizev1.Kustomization, error) {
	ksList := &kustomizev1.KustomizationList{}

	if err := kc.client.List(ctx, ksList); err != nil {
		return nil, fmt.Errorf("failed to list kustomizations: %w", err)
	}

	var kustomizations []kustomizev1.Kustomization
	for _, ks := range ksList.Items {
		if ks.Spec.Suspend {
			kustomizations = append(kustomizations, ks)
		}
	}

	return kustomizations, nil
}

func (kc *KubernetesClient) GetSuspendedHelmReleases(ctx context.Context) ([]helmv2.HelmRelease, error) {
	hrList := &helmv2.HelmReleaseList{}

	if err := kc.client.List(ctx, hrList); err != nil {
		return nil, fmt.Errorf("failed to list helm releases: %w", err)
	}

	var helmReleases []helmv2.HelmRelease
	for _, hr := range hrList.Items {
		if hr.Spec.Suspend {
			helmReleases = append(helmReleases, hr)
		}
	}

	return helmReleases, nil
}
