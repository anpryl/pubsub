let
  rev = "5c8c2587276506e61ec51e4b7d6c9288fd226794";
  src = builtins.fetchTarball
    "https://github.com/NixOS/nixpkgs/archive/${rev}.tar.gz";
  pkgs = import src {};

  pubsub-lint-run-tests = pkgs.writeScriptBin "pubsub-lint-run-tests" ''
    #!${pkgs.stdenv.shell}
    ${pkgs.golangci-lint}/bin/golangci-lint run && 
    ${pkgs.go}/bin/go test -v -timeout=30s --cover --race ./...
  '';
in 
  with pkgs; 
  pkgs.mkShell {
  buildInputs = [
    go
    dep
    golangci-lint
    pubsub-lint-run-tests
  ];
  shellHook = ''
    export GOROOT=${go.out}/share/go
  '';
}
