import subprocess
import os
import sys

def run_cmd(cmd, cwd=None):
    print(f"💻 Running: {cmd}")
    result = subprocess.run(cmd, shell=True, cwd=cwd)
    if result.returncode != 0:
        print(f"❌ Command failed: {cmd}")
        sys.exit(1)

def main():
    # Step 1: Install Golang
    print("📦 Installing Golang...")
    run_cmd("sudo apt update && sudo apt install -y golang")

    # Step 2: Clone the repository
    home_dir = os.path.expanduser("~")
    repo_dir = os.path.join(home_dir, "kuccps")

    if not os.path.exists(repo_dir):
        print("📁 Cloning repository...")
        run_cmd("git clone https://github.com/excreal/kuccps.git", cwd=home_dir)
    else:
        print("📁 Repository already cloned. Pulling latest changes...")
        run_cmd("git pull", cwd=repo_dir)

    # Step 3: Run build.sh
    print("🔨 Building the project...")
    run_cmd("bash build.sh", cwd=repo_dir)

    print("✅ Setup complete!")

if __name__ == "__main__":
    main()
