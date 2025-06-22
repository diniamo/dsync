{ buildGoModule, commit, lib, go }:
buildGoModule {
  pname = "dsync";
  version = "0-unstable-${commit}";

  src = lib.cleanSource ./.;
  
  vendorHash = "sha256-rVX3aSy9NkTTvhhysRLAUlS9C09ajUnelOAW273Y+Bw=";

  subPackages = [ "cmd/dsync" ];

  meta = {
    description = "Dead-simple P2P file synchronization tool using the SSH protocol";
    homepage = "https://github.com/diniamo/dsync";
    license = lib.licenses.eupl12;
    inherit (go.meta) platforms;
    maintainers = [lib.maintainers.diniamo];
    mainProgram = "rebuild";
  };
}
