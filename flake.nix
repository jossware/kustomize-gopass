{
  description = "kustomize-gopass: Generate secrets from gopass values.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

        app = pkgs.buildGo123Module {
          name = "kustomize-gopass";
          pname = "kustomize-gopass";

          src = ./.;

          # This hash locks the dependencies of this package. It is
          # necessary because of how Go requires network access to resolve
          # VCS.  See https://www.tweag.io/blog/2021-03-04-gomod2nix/ for
          # details. Normally one can build with a fake hash and rely on native Go
          # mechanisms to tell you what the hash should be or determine what
          # it should be "out-of-band" with other tooling (eg. gomod2nix).
          # To begin with it is recommended to set this, but one must
          # remember to bump this hash when your dependencies change.
          # vendorHash = pkgs.lib.fakeHash;
          vendorHash = "sha256-Vv42dLN0iKJUEL1o0ZxCHq741xufXHiakw97eVbkcxo=";
        };

      in
      with pkgs;
      {
        packages = {
          default = app;
        };
        devShells.default = mkShell {
          nativeBuildInputs = [
            go_1_23
            gopls
            gotools
            go-tools
            kustomize
            goreleaser
          ];
        };
        overlays = final: prev: { kustomize-gopass = app; };
      }
    );
}
