{ buildGoModule }:
buildGoModule {
  pname = "pductl";
  version = "0.1.0";
  src = ./.;
  vendorHash = "sha256-HoiYuyzHctwuwfEdySeMjdFe4juyqj+dPZ5U7OADdSc=";

  meta = {
    mainProgram = "pdud";
  };
}
