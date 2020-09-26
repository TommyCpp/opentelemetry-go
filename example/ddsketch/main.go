package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/metric"
	"go.opentelemetry.io/otel/exporters/stdout"
	array2 "go.opentelemetry.io/otel/sdk/metric/aggregator/array"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/ddsketch"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"log"
)

func initMeter() *basic.Processor {
	exp, _ := stdout.NewExporter(stdout.WithoutTraceExport())
	processor := basic.New(
		simple.NewWithSketchDistribution(ddsketch.NewDefaultConfig()),
		exp,
	)
	pusher := push.New(processor, exp)
	global.SetMeterProvider(pusher.Provider())
	return processor
}

func main() {
	pusher, err := stdout.InstallNewPipeline([]stdout.Option{
		stdout.WithQuantiles([]float64{0.5, 0.9, 0.99}),
		stdout.WithPrettyPrint(),
	}, nil)
	if err != nil {
		log.Fatalf("failed to initialize stdout export pipeline: %v", err)
	}
	defer pusher.Stop()

	descriptor := metric.NewDescriptor("test", metric.ValueRecorderKind, metric.Float64NumberKind)
	sketch := ddsketch.New(1, &descriptor, ddsketch.NewDefaultConfig())[0]
	array := array2.New(1)[0]
	for i := -1000.0; i < 100.0; i++ {
		_ = sketch.Update(context.Background(), metric.NewFloat64Number(i), &descriptor)
		_ = array.Update(context.Background(), metric.NewFloat64Number(i), &descriptor)
	}
	num, _ := sketch.Quantile(0.5)
	anum, _ := array.Quantile(0.5)
	fmt.Println(num.AsFloat64(), anum.AsFloat64())
	num, _ = sketch.Sum()
	anum, _ = array.Sum()
	fmt.Println(num.AsFloat64(), anum.AsFloat64())

}
