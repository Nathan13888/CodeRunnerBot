{
  description = "CodeRunnerBot flake";

  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem
      (
        system:
        let
          pkgs = import nixpkgs {
            inherit system;
          };
        in
        rec {
          packages.code-runner-bot = pkgs.buildGoModule rec {
            pname = "CodeRunnerBot";
            version = "0.1.0-dev";

            src = ./.;

            vendorSha256 = "sha256-SJK1c3Lf4hWoWTVLTTzMx1nudz7SZYT4uUluDuZ7K3E=";
          };
          defaultPackage = packages.code-runner-bot;

          devShell = pkgs.mkShell {
            nativeBuildInputs = with pkgs; [
              go
            ];
          };
        }
      );
}
