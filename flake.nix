{
  description = "Ceoris - Headless LMS & Community Management Platform";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Backend (Go)
            go
            gotools
            golangci-lint
            air # Live reloading para o Go

            # Frontend (Nuxt 4 / Node)
            nodejs_22
            pnpm
            
            # Utils
            sqlite
            just # Opcional: um command runner melhor que o Make
          ];

          shellHook = ''
            export CGO_ENABLED=1

            echo "🌌 Bem-vindo ao ambiente de desenvolvimento do Ceoris!"
            echo "🛠️  Go version: \$(go version)"
            echo "📦 Node version: \$(node -v)"
            echo "🚀 pnpm version: \$(pnpm -v)"
            echo ""
            echo "Comandos úteis:"
            echo "- Backend: cd server && air"
            echo "- Frontend: cd app && pnpm dev"
          '';
        };
      }
    );
}
