with import <nixpkgs> {};

stdenv.mkDerivation {

  name = "go-healthchecks";

  buildInputs = with pkgs; [
    jetbrains.goland
  ];

  shellHook = ''
    goland .
    exit
  '';
}

