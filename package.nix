{
  lib,
  stdenv,
  buildGoModule,
  fetchFromGitHub,
  installShellFiles,
  versionCheckHook,
}:

buildGoModule (finalAttrs: {
  pname = "omlox-client-go";
  version = "0.0.0";

  src = fetchFromGitHub {
    owner = "wavecomtech";
    repo = "omlox-client-go";
    tag = "v${finalAttrs.version}";
    hash = "";
  };

  vendorHash = "";

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${finalAttrs.version}"
    "-X main.commitHash=${finalAttrs.src.tag}"
    "-X main.commitDate=${finalAttrs.src.tag}" # src.lastModifiedDate works ?
  ];

  env.CGO_ENABLED = 0;

  nativeBuildInputs = [ installShellFiles ];
  postInstall = lib.optionalString (stdenv.buildPlatform.canExecute stdenv.hostPlatform) ''
    $out/bin/omlox-cli completion bash > omlox-cli.bash
    $out/bin/omlox-cli completion zsh > omlox-cli.zsh
    $out/bin/omlox-cli completion fish > omlox-cli.fish
    installShellCompletion omlox-cli.{bash,zsh,fish}
  '';

  nativeInstallCheckInputs = [ versionCheckHook ];
  doInstallCheck = true;

  meta = {
    description = "An Omlox Hub compatible Go client library and CLI tool";
    mainProgram = "omlox-cli";
    homepage = "https://github.com/wavecomtech/omlox-client-go";
    changelog = "https://github.com/wavecomtech/omlox-client-go/releases/tag/v${finalAttrs.version}";
    license = lib.licenses.mit;
    maintainers = with lib.maintainers; [ dvcorreia ];
  };
})
