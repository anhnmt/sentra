{ pkgs, lib, config, inputs, ... }:

{
  # https://devenv.sh/basics/

  # https://devenv.sh/packages/
  packages = [
  	pkgs.git
    pkgs.pkg-config
  	pkgs.yara-x
  	pkgs.yara
  ];

  # https://devenv.sh/languages/
	languages.go = {
		enable = true;
		version = "1.26.1";
	};

  # https://devenv.sh/processes/
  # processes.dev.exec = "${lib.getExe pkgs.watchexec} -n -- ls -la";

  # https://devenv.sh/services/
  # services.postgres.enable = true;

  # https://devenv.sh/scripts/
  scripts.init.exec = ''
  '';

  scripts.goget.exec = ''
		go get -u ./...
  '';

  scripts.tidy.exec = ''
		go mod tidy
  '';

  scripts.build-arm.exec = ''
		CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
		GOMAXPROCS=$(nproc) \
    go build \
      -trimpath \
      -ldflags "-extldflags '-static'" \
      -buildvcs=false \
      -o sentra
  '';

  scripts.build-amd.exec = ''
		CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
		GOMAXPROCS=$(nproc) \
    go build \
      -trimpath \
      -ldflags "-extldflags '-static'" \
      -buildvcs=false \
      -o sentra
  '';

  # https://devenv.sh/basics/
  enterShell = ''
    init         # Run scripts directly
  '';

  # https://devenv.sh/tasks/
  # tasks = {
  #   "myproj:setup".exec = "mytool build";
  #   "devenv:enterShell".after = [ "myproj:setup" ];
  # };

  # https://devenv.sh/tests/
  enterTest = ''
    echo "Running tests"
    git --version | grep --color=auto "${pkgs.git.version}"
  '';

  # https://devenv.sh/git-hooks/
  # git-hooks.hooks.shellcheck.enable = true;

  # See full reference at https://devenv.sh/reference/options/
}
