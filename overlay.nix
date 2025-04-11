final: prev: {
  omlox-client-go = final.callPackage ./package.nix { };
  omlox-client-go-nightly = final.callPackage ./nightly.nix { };
}
