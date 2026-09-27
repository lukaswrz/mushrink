{
  lib,
  buildGoModule,
}:
buildGoModule {
  pname = "mushrink";
  version = "0.0.0";

  src = lib.cleanSource ./.;

  vendorHash = "sha256-cZVg+LAQUkiC8PIhfJDumXq0sS+j9Bv1Kqy3wAr0mQs=";

  meta = {
    description = "A tool for compressing music";
    license = lib.licenses.gpl3Only;
    mainProgram = "mushrink";
  };
}
