{
  description = "A Go client for the Omlox Hub™";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      supportedSystems = [
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
        "x86_64-linux"
      ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f system);

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (
        system:
        import nixpkgs {
          inherit system;
          overlays = [ self.overlay ];
        }
      );
    in
    {
      overlay = import ./overlay.nix;

      formatter = forAllSystems (system: (nixpkgsFor.${system}).nixfmt-rfc-style);

      packages = forAllSystems (system: {
        default = (nixpkgsFor.${system}).omlox-client-go;
        omlox-client-go = (nixpkgsFor.${system}).omlox-client-go;
        omlox-client-go-nightly = (nixpkgsFor.${system}).omlox-client-go-nightly;
      });

      devShells = forAllSystems (
        system: with nixpkgsFor.${system}; {
          default = mkShell {
            inputsFrom = [ omlox-client-go-nightly ];
            packages = [
              git
              gnumake

              easyjson
              goreleaser
              copywrite
            ];
          };
        }
      );
    };
}
