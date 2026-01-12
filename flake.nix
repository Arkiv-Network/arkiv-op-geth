{
  description = "Arkiv OP Geth";
  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixos-unstable/nixexprs.tar.xz";

    systems.url = "github:nix-systems/default";

    rpcplorer = {
      # TODO: switch back to a release once the commit below was released
      url = "github:Golem-Base/rpcplorer?ref=1021adbfc765d9c36d565907cda7e5e48c6b597b";
      inputs = {
        systems.follows = "systems";
        nixpkgs.follows = "nixpkgs";
      };
    };
  };

  outputs =
    inputs:
    let
      eachSystem =
        f:
        inputs.nixpkgs.lib.genAttrs (import inputs.systems) (
          system: f system inputs.nixpkgs.legacyPackages.${system}
        );
    in
    {
      packages = eachSystem (
        _system: pkgs:
        let
          inherit (pkgs) lib;
        in
        {
          default = pkgs.buildGoModule {
            name = "gb-op-geth";

            src = ./.;

            subPackages = [
              "cmd/abidump"
              "cmd/abigen"
              "cmd/clef"
              "cmd/devp2p"
              "cmd/ethkey"
              "cmd/evm"
              "cmd/geth"
              "cmd/rlpdump"
              "cmd/utils"
            ];

            proxyVendor = true;
            vendorHash = "sha256-C8CCLMwN5MYIQ/3eQbVAJe5MPA8XgI9A2DFVVe4biDo=";

            ldflags = [
              "-s"
              "-w"
            ];

            meta = with lib; {
              description = "";
              homepage = "https://github.com/Golem-Base/golembase-op-geth";
              license = licenses.gpl3Only;
              mainProgram = "geth";
            };
          };

          golembase-cli = pkgs.buildGoModule {
            name = "golembase";
            src = ./.;
            subPackages = [ "cmd/golembase" ];
            vendorHash = "sha256-qkIg7Gks+4LvhlUR0+D6qjMQd7lcrybkbln4fPuAAKw=";
            meta = with lib; {
              description = "golembase CLI - Golem Base";
              homepage = "https://github.com/Golem-Base/golembase-op-geth";
              license = licenses.gpl3Only;
              mainProgram = "golembase";
            };
          };
        }
      );

      devShells = eachSystem (
        system: pkgs: {
          default = pkgs.mkShell {
            shellHook = ''
              # Set here the env vars you want to be available in the shell
            '';
            hardeningDisable = [ "all" ];

            packages =
              with pkgs;
              [
                go
                go-tools # staticccheck
                gopls # lsp
                gotools # goimports, ...
                shellcheck
                sqlc
                sqlite
                overmind
                mongosh
                openssl
                goreleaser
              ]
              ++ lib.optional pkgs.stdenv.hostPlatform.isLinux [
                # For podman networking
                slirp4netns
              ]
              ++ [ inputs.rpcplorer.packages.${system}.default ];
          };
        }
      );
    };
}
