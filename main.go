package main

import (
    "context"
    "flag"
    "log"

    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/DocSpring/terraform-provider-logstruct/pkg/provider"
)

// Set by goreleaser at build time
var (
    version = "dev"
)

func main() {
    var debug bool
    flag.BoolVar(&debug, "debug", false, "Enable debug mode.")
    flag.Parse()

    if err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
        Address: "github.com/DocSpring/logstruct",
        Debug:   debug,
    }); err != nil {
        log.Fatal(err)
    }
}

