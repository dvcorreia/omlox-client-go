{
  lib,
  stdenv,
  buildGoModule,
  fetchFromGitHub,
  installShellFiles,
}:

let
  fs = lib.fileset;
in
buildGoModule (finalAttrs: {
  pname = "omlox-client-go";
  version = "0.0.1-nightly";

  src = fs.toSource {
    root = ./.;
    fileset = fs.unions [
      ./go.mod
      ./go.sum
      (fs.fileFilter (file: lib.hasSuffix ".go" file.name) ./.)
    ];
  };

  vendorHash = "sha256-Gwkse7dl/EYxHSibgc6PMGwtDhKiMJYFajwAICpluzw=";

  ldflags = [
    "-X main.version=${finalAttrs.version}"
  ];

  env.CGO_ENABLED = 0;

  doInstallCheck = true;
  installCheckPhase = ''
    runHook preInstallCheck

    version_output=$($out/bin/omlox-cli version 2> /dev/null)
    if ! echo "$version_output" | grep -q "${finalAttrs.version}"; then
      echo "error: version '${finalAttrs.version}' not found in the command output"
      exit 1
    fi

    runHook postInstallCheck
  '';
})
