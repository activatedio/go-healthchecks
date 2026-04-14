with import <nixpkgs> {};

stdenv.mkDerivation {

  name = "go-healthchecks";

  buildInputs = with pkgs; [
    go
    gnumake
  ];

  shellHook = ''
    export GOPATH=$HOME/go
    export PATH=$PATH:$HOME/go/bin
  '';

  hardeningDisable = [ "fortify" ];

}


