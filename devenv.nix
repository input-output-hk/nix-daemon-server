{pkgs, ...}: {
  # https://devenv.sh/basics/
  env.GITHUB_ORGANIZATION = "input-output-hk";
  env.GITHUB_TEAM = "devops";
  env.PGDATA = ".pgdata";

  # https://devenv.sh/packages/
  packages = with pkgs; [git gopls watchexec golangci-lint dbmate gcc];
  postgres.enable = true;

  enterShell = ''
    export PS1="\[\033[1;32m\][\[\e]0;\u@\h: \w\a\]\u@\h:\w]\$\[\033[0m\] "
    export DATABASE_URL="postgres://$USER/$USER?host=$PWD/.pgdata&sslmode=disable&search_path=$USER";
  '';

  # https://devenv.sh/languages/
  languages.nix.enable = true;
  languages.go.enable = true;

  # https://devenv.sh/scripts/
  scripts.lint.exec = "golangci-lint run";
  scripts.run.exec = "go run ./pkg/nix-daemon-server";

  # https://devenv.sh/pre-commit-hooks/
  pre-commit.hooks = {
    shellcheck.enable = true;
    alejandra.enable = true;
    deadnix.enable = true;
    # govet.enable = true;
    # nix-linter.enable = true;
    # revive.enable = true;
    # statix.enable = true;
  };

  # https://devenv.sh/processes/
  processes.watch.exec = "go run ./pkg/nix-daemon-server";
}
