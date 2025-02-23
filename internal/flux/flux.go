package flux

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.Meter("alpineworks.io/flux-suspension-exporter/internal/flux")
)

var (
	TypeKey                    = "type"
	TypeHelmReleaseAttribute   = otelmetric.WithAttributes(attribute.String(TypeKey, "helm-release"))
	TypeKustomizationAttribute = otelmetric.WithAttributes(attribute.String(TypeKey, "kustomization"))

	NameKey       = "name"
	NameAttribute = func(name string) otelmetric.MeasurementOption {
		return otelmetric.WithAttributes(attribute.String(NameKey, name))
	}

	NamespaceKey       = "namespace"
	NamespaceAttribute = func(namespace string) otelmetric.MeasurementOption {
		return otelmetric.WithAttributes(attribute.String(NamespaceKey, namespace))
	}
)

func NewMetrics(kc *KubernetesClient) error {
	_, err := meter.Int64ObservableGauge(
		"flux.suspended",
		otelmetric.WithDescription("The suspended fluxcd resources in the kubernetes cluster"),
		otelmetric.WithInt64Callback(kc.FluxSuspensionGauge()),
	)

	if err != nil {
		return fmt.Errorf("failed to create int64 observable gauge: %s", err)
	}

	return nil
}

func (kc *KubernetesClient) FluxSuspensionGauge() otelmetric.Int64Callback {
	return func(ctx context.Context, i64Observer otelmetric.Int64Observer) error {
		ksCtx, ksCancel := context.WithTimeout(ctx, time.Second*5)
		defer ksCancel()

		kustomizations, err := kc.GetSuspendedKustomizations(ksCtx)
		if err != nil {
			slog.Error("failed to get suspended kustomizations", slog.String("error", err.Error()))
		}

		hrCtx, hrCancel := context.WithTimeout(ctx, time.Second*5)
		defer hrCancel()

		helmReleases, err := kc.GetSuspendedHelmReleases(hrCtx)
		if err != nil {
			slog.Error("failed to get suspended helmreleases", slog.String("error", err.Error()))
		}

		for _, ks := range kustomizations {
			i64Observer.Observe(1, NameAttribute(ks.Name), NamespaceAttribute(ks.Namespace), TypeKustomizationAttribute)
		}

		for _, hr := range helmReleases {
			i64Observer.Observe(1, NameAttribute(hr.Name), NamespaceAttribute(hr.Namespace), TypeHelmReleaseAttribute)
		}

		return nil
	}

}
