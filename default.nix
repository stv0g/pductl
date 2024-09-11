{ buildGoModule }:
buildGoModule {
  pname = "pductl";
  version = "0.1.0";
  src = ./.;
  vendorHash = "sha256-RnwlwVwU1O2yYMPtUbyz8Xqv1gIZdKNapNhd9QnLjHk=";

  meta = {
    mainProgram = "pdud";
  };
}
