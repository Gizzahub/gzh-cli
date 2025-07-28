class GzhGitExtensions < Formula
  desc "Git extensions for enhanced repository management"
  homepage "https://github.com/gizzahub/gzh-manager-go"
  url "https://github.com/gizzahub/gzh-manager-go/archive/v1.0.0.tar.gz"
  sha256 "sha256-placeholder" # This will be updated when creating actual releases
  license "MIT"
  head "https://github.com/gizzahub/gzh-manager-go.git", branch: "master"

  depends_on "go" => :build

  def install
    # Set version information
    version_flags = [
      "-X main.Version=#{version}",
      "-X main.BuildTime=#{Time.now.utc.strftime("%Y-%m-%d_%H:%M:%S")}",
    ]
    
    # Build git-synclone
    system "go", "build", "-ldflags", version_flags.join(" "), "-o", "git-synclone", "./cmd/git-synclone"
    
    # Install the binary
    bin.install "git-synclone"
  end

  test do
    # Test that the binary was installed and runs
    system "#{bin}/git-synclone", "--version"
    
    # Test that git integration works
    system "git", "synclone", "--help"
  end

  def caveats
    <<~EOS
      git-synclone has been installed as a Git extension.
      
      You can now use:
        git synclone --help
        git synclone github -o myorg
        git synclone gitlab -g mygroup
        git synclone gitea -o myorg
      
      Configuration file can be placed at:
        ~/.config/gzh-manager/synclone.yaml
      
      For more information and examples:
        https://github.com/gizzahub/gzh-manager-go
    EOS
  end
end