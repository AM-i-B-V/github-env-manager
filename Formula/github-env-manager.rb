class GithubEnvManager < Formula
  desc "A modern, web-based tool for managing GitHub repository variables and secrets across environments"
  homepage "https://github.com/AM-i-B-V/github-env-manager"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/AM-i-B-V/github-env-manager/releases/download/v0.1.0/github-env-manager_darwin_arm64.tar.gz"
      sha256 "YOUR_SHA256_HERE" # Replace with actual SHA256
    else
      url "https://github.com/AM-i-B-V/github-env-manager/releases/download/v0.1.0/github-env-manager_darwin_amd64.tar.gz"
      sha256 "YOUR_SHA256_HERE" # Replace with actual SHA256
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/AM-i-B-V/github-env-manager/releases/download/v0.1.0/github-env-manager_linux_arm64.tar.gz"
      sha256 "YOUR_SHA256_HERE" # Replace with actual SHA256
    else
      url "https://github.com/AM-i-B-V/github-env-manager/releases/download/v0.1.0/github-env-manager_linux_amd64.tar.gz"
      sha256 "YOUR_SHA256_HERE" # Replace with actual SHA256
    end
  end

  def install
    bin.install "github-env-manager"
  end

  test do
    system "#{bin}/github-env-manager", "--help"
  end
end
