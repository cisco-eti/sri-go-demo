{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/432fc2d9a67f92e05438dff5fdc2b39d33f77997.tar.gz") {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.git
    pkgs.go_1_16
  ];

  GOPRIVATE = "wwwin-github.cisco.com";

  shellHook = ''
    echo -n Password:
    read -s pass

    export GOPROXY="https://proxy.golang.org, https://$(whoami):$pass@engci-maven-master.cisco.com/artifactory/api/go/nyota-go, direct"
  '';
}
