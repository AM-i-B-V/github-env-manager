// GitHub Environment Manager - Modern UI (React-inspired)

class GitHubEnvManager {
  constructor() {
    this.user = null;
    this.sessionId = null;
    this.ownerRepo = { owner: "keyvalue", name: "payments-svc" };
    this.envs = [];
    this.selectedEnvs = [];
    this.targetEnvs = [];
    this.metas = {};
    this.exporting = false;
    this.exportText = "";
    this.importTargets = [];
    this.importPreview = {};
    this.activeScopeTab = "repo"; // 'repo' | 'org'

    this.init();
  }

  init() {
    this.bindEvents();
    this.checkAuthStatus();
  }

  bindEvents() {
    // Authentication
    document
      .getElementById("authBtn")
      .addEventListener("click", () => this.authenticate());
    document
      .getElementById("logoutBtn")
      .addEventListener("click", () => this.logout());

    // Repository management
    document
      .getElementById("searchReposBtn")
      .addEventListener("click", () => this.searchRepositories());
    document.getElementById("repoSearch").addEventListener("keypress", (e) => {
      if (e.key === "Enter") {
        this.searchRepositories();
      }
    });

    // Environment management
    document
      .getElementById("createEnvBtn")
      .addEventListener("click", () => this.showCreateEnvModal());
    document
      .getElementById("refreshBtn")
      .addEventListener("click", () => this.refreshMeta());

    // Add key form
    document
      .getElementById("addKeyForm")
      .addEventListener("submit", (e) => this.handleAddKey(e));

    // Export/Import
    document
      .getElementById("importFile")
      .addEventListener("change", (e) => this.handleImportFile(e));
    document
      .getElementById("importText")
      .addEventListener("input", (e) => this.handleImportText(e));
    document
      .getElementById("applyImportBtn")
      .addEventListener("click", () => this.applyImport());
    document
      .getElementById("clearImportBtn")
      .addEventListener("click", () => this.clearImport());

    // Repository scope
    document
      .getElementById("repoScopeBtn")
      .addEventListener("click", () => this.loadRepoScopeData());

    // Repo scope
    document
      .getElementById("addRepoKeyBtn")
      .addEventListener("click", () => this.addRepoKey());

    // Token form
    document
      .getElementById("tokenForm")
      .addEventListener("submit", (e) => this.handleTokenSubmit(e));

    // Create environment form
    document
      .getElementById("createEnvForm")
      .addEventListener("submit", (e) => this.handleCreateEnvSubmit(e));
  }

  async checkAuthStatus() {
    // Check if we have stored authentication
    const storedToken = localStorage.getItem("github_token");
    const storedUser = localStorage.getItem("github_user");

    if (storedToken && storedUser) {
      try {
        // Try to validate the stored token
        const response = await fetch("/api/auth/validate", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ token: storedToken }),
        });

        if (response.ok) {
          const data = await response.json();
          this.sessionId = data.sessionId;
          this.user = JSON.parse(storedUser);
          this.showUserInfo();
          this.loadRepositories();
          return;
        }
      } catch (error) {
        console.error("Stored token validation failed:", error);
      }

      // If validation failed, clear stored data
      localStorage.removeItem("github_token");
      localStorage.removeItem("github_user");
    }

    this.showAuthSection();
  }

  async loadRepositories(page = 1) {
    try {
      const response = await fetch(`/api/repos?page=${page}&per_page=25`, {
        headers: {
          "X-Session-ID": this.sessionId || "",
        },
      });
      if (response.ok) {
        const data = await response.json();

        // Handle both new paginated format and old format for backward compatibility
        let repos, pagination;
        if (data.repositories && data.pagination) {
          // New paginated format
          repos = data.repositories;
          pagination = data.pagination;
        } else if (Array.isArray(data)) {
          // Old format (fallback)
          repos = data;
          pagination = null;
        } else {
          // Unexpected format
          repos = [];
          pagination = null;
        }

        this.renderRepositoryList(repos, pagination);
        this.showRepositorySection();
      } else {
        this.showToast("Failed to load repositories", "error");
      }
    } catch (error) {
      this.showToast("Failed to load repositories", "error");
      console.error("Load repositories error:", error);
    }
  }

  showRepositorySection() {
    // The repoPicker is already visible by default, just ensure mainContent is hidden
    const mainContent = document.getElementById("mainContent");
    if (mainContent) {
      mainContent.classList.add("hidden");
    }
  }

  async authenticate() {
    // Show token input modal
    this.showTokenModal();
  }

  showTokenModal() {
    document.getElementById("tokenModal").classList.remove("hidden");
  }

  closeTokenModal() {
    document.getElementById("tokenModal").classList.add("hidden");
    document.getElementById("tokenInput").value = "";
  }

  showCreateEnvModal() {
    document.getElementById("createEnvModal").classList.remove("hidden");
    document.getElementById("newEnvName").focus();
  }

  closeCreateEnvModal() {
    document.getElementById("createEnvModal").classList.add("hidden");
    document.getElementById("newEnvName").value = "";
    document.getElementById("newEnvDescription").value = "";
  }

  async authenticateWithToken(token) {
    try {
      const response = await fetch("/api/auth/pat", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ token }),
      });

      if (response.ok) {
        const data = await response.json();
        this.sessionId = data.sessionId;
        this.user = data.user;

        // Check if "Remember Me" is checked
        const rememberMe = document.getElementById("rememberMe").checked;
        if (rememberMe) {
          // Store authentication data for persistence
          localStorage.setItem("github_token", token);
          localStorage.setItem("github_user", JSON.stringify(data.user));
          this.showToast(
            "Authentication successful - Token saved for future sessions",
            "success"
          );
        } else {
          // Clear any previously stored data
          localStorage.removeItem("github_token");
          localStorage.removeItem("github_user");
          this.showToast(
            "Authentication successful - Token not saved",
            "success"
          );
        }

        this.closeTokenModal();
        this.showUserInfo();
        this.loadRepositories();
      } else {
        const error = await response.json();
        this.showToast(error.error || "Authentication failed", "error");
      }
    } catch (error) {
      this.showToast("Authentication failed", "error");
      console.error("Authentication error:", error);
    }
  }

  async handleTokenSubmit(e) {
    e.preventDefault();
    const token = document.getElementById("tokenInput").value.trim();
    if (!token) {
      this.showToast("Please enter a token", "error");
      return;
    }
    await this.authenticateWithToken(token);
  }

  async handleCreateEnvSubmit(e) {
    e.preventDefault();

    const name = document.getElementById("newEnvName").value.trim();
    const description = document
      .getElementById("newEnvDescription")
      .value.trim();

    if (!name) {
      this.showToast("Environment name is required", "error");
      return;
    }

    // Validate environment name format
    if (!/^[a-z0-9-]+$/.test(name)) {
      this.showToast(
        "Environment name can only contain lowercase letters, numbers, and hyphens",
        "error"
      );
      return;
    }

    // Check if environment already exists
    if (this.envs.includes(name)) {
      this.showToast(`Environment "${name}" already exists`, "error");
      return;
    }

    this.showLoading(true);
    try {
      await this.createEnvironment(name, description);
      this.closeCreateEnvModal();
      this.showToast(`Environment "${name}" created successfully`, "success");

      // Refresh environments list
      await this.loadEnvironments();
    } catch (error) {
      this.showToast("Failed to create environment", "error");
      console.error("Create environment error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  async createEnvironment(name, description) {
    const response = await fetch(
      `/api/repos/${this.ownerRepo.owner}/${this.ownerRepo.name}/environments`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Session-ID": this.sessionId || "",
        },
        body: JSON.stringify({
          name,
          description,
        }),
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to create environment");
    }

    return response.json();
  }

  logout() {
    this.user = null;
    this.sessionId = null;

    // Clear stored authentication data
    localStorage.removeItem("github_token");
    localStorage.removeItem("github_user");

    this.showAuthSection();
    this.showToast("Logged out successfully", "success");
  }

  showUserInfo() {
    const userInfo = document.getElementById("userInfo");
    const authBtn = document.getElementById("authBtn");
    const userAvatar = document.getElementById("userAvatar");
    const userName = document.getElementById("userName");

    if (userInfo) userInfo.classList.remove("hidden");
    if (authBtn) authBtn.classList.add("hidden");
    if (userAvatar) userAvatar.src = this.user.avatarUrl || "";
    if (userName) {
      const storedToken = localStorage.getItem("github_token");
      const indicator = storedToken ? " 🔐" : "";
      userName.textContent = (this.user.login || "") + indicator;
    }
  }

  showAuthSection() {
    const userInfo = document.getElementById("userInfo");
    const authBtn = document.getElementById("authBtn");
    const mainContent = document.getElementById("mainContent");

    if (userInfo) userInfo.classList.add("hidden");
    if (authBtn) authBtn.classList.remove("hidden");
    if (mainContent) mainContent.classList.add("hidden");
  }

  async searchRepositories(page = 1) {
    const query = document.getElementById("repoSearch").value;
    this.showLoading(true);

    try {
      const response = await fetch(
        `/api/repos?q=${encodeURIComponent(query)}&page=${page}&per_page=25`,
        {
          headers: {
            "X-Session-ID": this.sessionId || "",
          },
        }
      );
      if (response.ok) {
        const data = await response.json();

        // Handle both new paginated format and old format for backward compatibility
        let repos, pagination;
        if (data.repositories && data.pagination) {
          // New paginated format
          repos = data.repositories;
          pagination = data.pagination;
        } else if (Array.isArray(data)) {
          // Old format (fallback)
          repos = data;
          pagination = null;
        } else {
          // Unexpected format
          repos = [];
          pagination = null;
        }

        this.renderRepositoryList(repos, pagination);

        // Show search results count
        if (query) {
          this.showToast(
            `Found ${repos.length} repositories matching "${query}"`,
            "success"
          );
        } else {
          this.showToast(
            `Showing ${repos.length} of your accessible repositories`,
            "success"
          );
        }
      } else {
        const error = await response.json();
        this.showToast(error.error || "Failed to search repositories", "error");
      }
    } catch (error) {
      this.showToast("Failed to search repositories", "error");
      console.error("Search repositories error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  renderRepositoryList(repos, pagination = null) {
    const repoList = document.getElementById("repoList");
    const query = document.getElementById("repoSearch").value;
    repoList.innerHTML = "";

    if (repos.length === 0) {
      repoList.innerHTML = `
        <div class="col-span-full text-center py-8 text-neutral-500">
          <div class="text-sm">No repositories found</div>
          <div class="text-xs mt-1">
            ${
              query
                ? `Try a different search term or check your permissions`
                : `Loading repositories...`
            }
          </div>
          <div class="text-xs mt-2 text-blue-600">
            <i class="fas fa-info-circle mr-1"></i>
            Tip: Try searching by repository name, organization, or description
          </div>
        </div>
      `;
      return;
    }

    // Add a note about pagination
    if (!pagination || pagination.page === 1) {
      const noteDiv = document.createElement("div");
      noteDiv.className =
        "col-span-full mb-4 p-3 bg-blue-50 border border-blue-200 rounded-lg";
      noteDiv.innerHTML = `
        <div class="text-sm text-blue-800">
          <i class="fas fa-info-circle mr-2"></i>
          Showing up to 25 repositories per page. Use pagination controls below to browse more repositories.
        </div>
      `;
      repoList.appendChild(noteDiv);
    }

    repos.forEach((repo) => {
      const button = document.createElement("button");
      button.className =
        "flex items-center justify-between rounded-xl border p-3 text-left hover:bg-neutral-50";

      const visibilityIcon = repo.private ? "🔒" : "🌐";
      const description = repo.description || "No description";

      button.innerHTML = `
        <div class="flex-1">
          <div class="flex items-center gap-2">
            <div class="text-sm font-semibold">${
              repo.owner.login || repo.owner
            }/${repo.name}</div>
            <span class="text-xs">${visibilityIcon}</span>
          </div>
          <div class="text-xs text-neutral-500 mt-1">${description}</div>
        </div>
        <span class="inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium">Pick</span>
      `;

      button.addEventListener("click", () => this.selectRepository(repo));
      repoList.appendChild(button);
    });

    // Add pagination controls if pagination info is available
    if (pagination) {
      const paginationDiv = document.createElement("div");
      paginationDiv.className =
        "col-span-full flex items-center justify-between mt-6 p-4 bg-slate-50 rounded-lg";

      const infoText = document.createElement("div");
      infoText.className = "text-sm text-slate-600";
      const totalPages = pagination.total_pages || "?";
      infoText.textContent = `Page ${pagination.page} of ${totalPages} (${repos.length} repositories)`;

      const controlsDiv = document.createElement("div");
      controlsDiv.className = "flex items-center gap-2";

      // Previous button
      if (pagination.has_prev) {
        const prevBtn = document.createElement("button");
        prevBtn.className =
          "px-3 py-1.5 rounded-lg text-sm font-medium bg-white border border-slate-300 text-slate-700 hover:bg-slate-50 transition-colors";
        prevBtn.innerHTML =
          '<i class="fas fa-chevron-left text-xs mr-1"></i>Previous';
        prevBtn.addEventListener("click", () =>
          this.searchRepositories(pagination.page - 1)
        );
        controlsDiv.appendChild(prevBtn);
      }

      // Next button
      if (pagination.has_next) {
        const nextBtn = document.createElement("button");
        nextBtn.className =
          "px-3 py-1.5 rounded-lg text-sm font-medium bg-white border border-slate-300 text-slate-700 hover:bg-slate-50 transition-colors";
        nextBtn.innerHTML =
          'Next<i class="fas fa-chevron-right text-xs ml-1"></i>';
        nextBtn.addEventListener("click", () =>
          this.searchRepositories(pagination.page + 1)
        );
        controlsDiv.appendChild(nextBtn);
      }

      paginationDiv.appendChild(infoText);
      paginationDiv.appendChild(controlsDiv);
      repoList.appendChild(paginationDiv);
    }
  }

  selectRepository(repo) {
    this.ownerRepo = {
      owner: repo.owner.login || repo.owner,
      name: repo.name,
    };
    this.selectedEnvs = [];
    this.importTargets = [];
    this.exportText = "";
    this.importPreview = {};

    this.updateContext();
    this.loadEnvironments();
    this.showMainContent();
  }

  updateOwnerRepo(updates) {
    this.ownerRepo = { ...this.ownerRepo, ...updates };
    this.updateContext();
  }

  updateContext() {
    const repoContext = document.getElementById("repoContext");
    const repoUrl = document.getElementById("repoUrl");

    if (this.ownerRepo.owner && this.ownerRepo.name) {
      repoContext.textContent = `${this.ownerRepo.owner}/${this.ownerRepo.name}`;
      const url = `https://github.com/${this.ownerRepo.owner}/${this.ownerRepo.name}`;
      repoUrl.textContent = url;
      repoUrl.href = url;
      repoUrl.classList.remove("hidden");
    } else {
      repoContext.textContent = "Select a repository";
      repoUrl.classList.add("hidden");
    }

    document.getElementById("envContext").textContent = this.selectedEnvs.length
      ? this.selectedEnvs.join(" · ")
      : "No environments selected";

    this.updateStatsCard();
  }

  updateStatsCard() {
    const statsContent = document.getElementById("statsContent");

    if (!this.selectedEnvs.length || !Object.keys(this.metas).length) {
      statsContent.textContent = "No data";
      return;
    }

    let totalVariables = 0;
    let totalSecrets = 0;

    this.selectedEnvs.forEach((env) => {
      const meta = this.metas[env] || { variables: {}, secrets: {} };
      totalVariables += Object.keys(meta.variables).length;
      totalSecrets += Object.keys(meta.secrets).length;
    });

    statsContent.innerHTML = `
      <div class="flex items-center gap-2">
        <span class="inline-flex items-center px-2 py-1 bg-blue-100 text-blue-800 rounded-md text-xs font-medium">
          <i class="fas fa-wrench text-xs mr-1"></i>
          ${totalVariables} vars
        </span>
        <span class="inline-flex items-center px-2 py-1 bg-orange-100 text-orange-800 rounded-md text-xs font-medium">
          <i class="fas fa-lock text-xs mr-1"></i>
          ${totalSecrets} secrets
        </span>
      </div>
    `;
  }

  showMainContent() {
    const mainContent = document.getElementById("mainContent");
    if (mainContent) {
      mainContent.classList.remove("hidden");
    }
  }

  async loadEnvironments() {
    if (!this.ownerRepo.owner || !this.ownerRepo.name) return;

    this.showLoading(true);
    try {
      const response = await fetch(
        `/api/repos/${this.ownerRepo.owner}/${this.ownerRepo.name}/environments`,
        {
          headers: {
            "X-Session-ID": this.sessionId || "",
          },
        }
      );
      if (response.ok) {
        this.envs = await response.json();
        this.selectedEnvs = []; // Start with no environments selected
        this.importTargets = this.envs;
        this.renderEnvironmentPicker();
        this.renderTargetEnvs();
        this.renderImportTargets();
        this.renderExportButtons();
        this.updateContext();
        this.loadMeta();
        this.loadRepoScopeData(); // Also load repository variables and secrets
      } else {
        const error = await response.json();
        this.showToast(error.error || "Failed to load environments", "error");
      }
    } catch (error) {
      this.showToast("Failed to load environments", "error");
      console.error("Load environments error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  renderEnvironmentPicker() {
    const envPicker = document.getElementById("envPicker");
    envPicker.innerHTML = "";

    if (this.envs.length === 0) {
      envPicker.innerHTML = `
        <div class="p-4 text-center text-neutral-500 bg-neutral-50 rounded-xl border border-neutral-200">
          <i class="fas fa-info-circle text-neutral-400 mb-2"></i>
          <div class="text-sm">No environments found for this repository</div>
          <div class="text-xs mt-1">Create environments to manage variables and secrets</div>
        </div>
      `;
      return;
    }

    // Always show the control section
    const controlSection = document.createElement("div");
    controlSection.className =
      "mb-4 p-4 bg-slate-50 rounded-xl border border-slate-200";

    if (this.selectedEnvs.length === 0) {
      controlSection.innerHTML = `
        <div class="flex items-center justify-between">
          <div class="flex-1">
            <div class="text-sm font-medium text-slate-700 mb-1">Select Environments</div>
            <div class="text-xs text-slate-500">Choose environments to compare variables and secrets</div>
          </div>
          <button class="px-4 py-2 rounded-lg text-sm font-medium shadow-sm border bg-blue-600 text-white hover:bg-blue-700 transition-colors" onclick="app.selectAllEnvironments()">
            <i class="fas fa-check-double mr-2"></i>Select All
          </button>
        </div>
      `;
    } else {
      const selectedCount = this.selectedEnvs.length;
      const totalCount = this.envs.length;
      controlSection.innerHTML = `
        <div class="flex items-center justify-between">
          <div class="flex-1">
            <div class="text-sm font-medium text-slate-700 mb-1">Selected Environments</div>
            <div class="text-xs text-slate-500">${selectedCount} of ${totalCount} environments selected</div>
          </div>
          <div class="flex items-center gap-2">
            <button class="px-3 py-2 rounded-lg text-sm font-medium shadow-sm border bg-blue-600 text-white hover:bg-blue-700 transition-colors" onclick="app.selectAllEnvironments()">
              <i class="fas fa-check-double mr-1"></i>Select All
            </button>
            <button class="px-3 py-2 rounded-lg text-sm font-medium shadow-sm border bg-gray-600 text-white hover:bg-gray-700 transition-colors" onclick="app.deselectAllEnvironments()">
              <i class="fas fa-times mr-1"></i>Clear All
            </button>
          </div>
        </div>
      `;
    }

    envPicker.appendChild(controlSection);

    // Environment buttons
    const envButtonsSection = document.createElement("div");
    envButtonsSection.className =
      "grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3";

    this.envs.forEach((env) => {
      const button = document.createElement("button");
      button.type = "button"; // Prevent form submission
      button.className = `flex items-center justify-between rounded-xl border px-4 py-3 text-sm cursor-pointer hover:bg-neutral-50 transition-all duration-200 ${
        this.selectedEnvs.includes(env)
          ? "bg-blue-600 text-white border-blue-600 shadow-md"
          : "bg-white border-gray-300 hover:border-gray-400"
      }`;
      button.innerHTML = `
        <div class="flex items-center gap-3">
          <i class="fas ${
            this.selectedEnvs.includes(env) ? "fa-check-circle" : "fa-circle"
          } text-sm"></i>
          <span class="font-medium">${env}</span>
        </div>
        <div class="text-xs opacity-75">
          ${this.selectedEnvs.includes(env) ? "Selected" : "Click to select"}
        </div>
      `;

      button.addEventListener("click", (e) => {
        e.preventDefault();
        e.stopPropagation();
        console.log(
          "Environment clicked:",
          env,
          "Current selected:",
          this.selectedEnvs
        );

        // Add a brief visual feedback
        button.style.transform = "scale(0.98)";
        setTimeout(() => {
          button.style.transform = "";
        }, 150);

        this.toggleEnvironment(env);
      });
      envButtonsSection.appendChild(button);
    });

    envPicker.appendChild(envButtonsSection);
  }

  toggleEnvironment(env) {
    console.log("toggleEnvironment called with:", env);
    console.log("Before toggle - selectedEnvs:", this.selectedEnvs);

    if (this.selectedEnvs.includes(env)) {
      this.selectedEnvs = this.selectedEnvs.filter((e) => e !== env);
      // Also remove from target environments if it was selected there
      this.targetEnvs = this.targetEnvs.filter((e) => e !== env);
      console.log("Removed environment:", env);
    } else {
      this.selectedEnvs = [...this.selectedEnvs, env];
      console.log("Added environment:", env);
    }

    console.log("After toggle - selectedEnvs:", this.selectedEnvs);

    // Update the UI
    this.renderEnvironmentPicker();
    this.renderTargetEnvs();
    this.renderExportButtons();
    this.updateContext();

    // Only load meta if there are selected environments
    if (this.selectedEnvs.length > 0) {
      this.loadMeta();
    } else {
      this.renderCompareTable(); // Show empty state
    }
  }

  selectAllEnvironments() {
    this.selectedEnvs = [...this.envs];
    this.renderEnvironmentPicker();
    this.renderTargetEnvs();
    this.renderExportButtons();
    this.updateContext();
    this.loadMeta();
  }

  deselectAllEnvironments() {
    this.selectedEnvs = [];
    this.targetEnvs = []; // Also clear target environments
    this.renderEnvironmentPicker();
    this.renderTargetEnvs();
    this.renderExportButtons();
    this.updateContext();
    this.renderCompareTable(); // Clear the comparison table
  }

  renderTargetEnvs() {
    const targetEnvs = document.getElementById("targetEnvs");
    targetEnvs.innerHTML = "";

    if (this.selectedEnvs.length === 0) {
      targetEnvs.innerHTML = `
        <div class="p-3 text-center text-neutral-500 bg-neutral-50 rounded-lg border border-neutral-200">
          <i class="fas fa-info-circle text-neutral-400 mb-1"></i>
          <div class="text-xs">Select environments first to add variables/secrets</div>
        </div>
      `;
      return;
    }

    this.selectedEnvs.forEach((env) => {
      const label = document.createElement("label");
      label.className = `flex items-center justify-between rounded-lg border px-4 py-3 text-sm cursor-pointer hover:bg-neutral-50 transition-all duration-200 ${
        this.targetEnvs.includes(env)
          ? "bg-blue-600 text-white border-blue-600 shadow-sm"
          : "bg-white border-gray-300 hover:border-gray-400"
      }`;
      label.innerHTML = `
        <div class="flex items-center gap-3">
          <input type="checkbox" ${
            this.targetEnvs.includes(env) ? "checked" : ""
          } class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500">
          <span class="font-medium">${env}</span>
        </div>
        <div class="text-xs opacity-75">
          ${this.targetEnvs.includes(env) ? "Selected" : "Click to select"}
        </div>
      `;

      label.addEventListener("click", () => this.toggleTargetEnv(env));
      targetEnvs.appendChild(label);
    });
  }

  renderImportTargets() {
    const importTargets = document.getElementById("importTargets");
    importTargets.innerHTML = "";

    if (this.envs.length === 0) {
      importTargets.innerHTML = `
        <div class="p-3 text-center text-neutral-500 bg-neutral-50 rounded-lg border border-neutral-200">
          <i class="fas fa-info-circle text-neutral-400 mb-1"></i>
          <div class="text-xs">No environments available for import</div>
        </div>
      `;
      return;
    }

    this.envs.forEach((env) => {
      const label = document.createElement("label");
      label.className = `flex items-center justify-between rounded-lg border px-4 py-3 text-sm cursor-pointer hover:bg-neutral-50 transition-all duration-200 ${
        this.importTargets.includes(env)
          ? "bg-blue-600 text-white border-blue-600 shadow-sm"
          : "bg-white border-gray-300 hover:border-gray-400"
      }`;
      label.innerHTML = `
        <div class="flex items-center gap-3">
          <input type="checkbox" ${
            this.importTargets.includes(env) ? "checked" : ""
          } class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500">
          <span class="font-medium">${env}</span>
        </div>
        <div class="text-xs opacity-75">
          ${this.importTargets.includes(env) ? "Selected" : "Click to select"}
        </div>
      `;

      label.addEventListener("click", () => this.toggleImportTarget(env));
      importTargets.appendChild(label);
    });
  }

  toggleImportTarget(env) {
    if (this.importTargets.includes(env)) {
      this.importTargets = this.importTargets.filter((e) => e !== env);
    } else {
      this.importTargets = [...this.importTargets, env];
    }
    this.renderImportTargets();
  }

  toggleTargetEnv(env) {
    if (this.targetEnvs.includes(env)) {
      this.targetEnvs = this.targetEnvs.filter((e) => e !== env);
    } else {
      this.targetEnvs = [...this.targetEnvs, env];
    }
    this.renderTargetEnvs();
  }

  renderExportButtons() {
    const exportButtons = document.getElementById("exportButtons");
    exportButtons.innerHTML = "";

    if (this.selectedEnvs.length === 0) {
      exportButtons.innerHTML = `
        <div class="p-3 text-center text-neutral-500 bg-neutral-50 rounded-lg border border-neutral-200">
          <i class="fas fa-info-circle text-neutral-400 mb-1"></i>
          <div class="text-xs">Select environments to export variables</div>
        </div>
      `;
      return;
    }

    this.selectedEnvs.forEach((env) => {
      const button = document.createElement("button");
      button.className =
        "flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium shadow-sm border bg-white hover:bg-neutral-50 transition-colors";
      button.innerHTML = `
        <i class="fas fa-download text-xs"></i>
        <span>Export ${env}.env</span>
      `;
      button.addEventListener("click", () => this.exportEnvironment(env));
      exportButtons.appendChild(button);
    });
  }

  async loadMeta() {
    if (!this.selectedEnvs.length) {
      this.renderCompareTable(); // Clear the table when no environments selected
      return;
    }

    this.showLoading(true);
    try {
      // Load environment-specific variables and secrets for each environment
      const promises = this.selectedEnvs.map(async (env) => {
        const variablesResponse = await fetch(
          `/api/repos/${this.ownerRepo.owner}/${this.ownerRepo.name}/environments/${env}/variables`,
          {
            headers: {
              "X-Session-ID": this.sessionId || "",
            },
          }
        );
        const variables = variablesResponse.ok
          ? await variablesResponse.json()
          : [];

        const secretsResponse = await fetch(
          `/api/repos/${this.ownerRepo.owner}/${this.ownerRepo.name}/environments/${env}/secrets`,
          {
            headers: {
              "X-Session-ID": this.sessionId || "",
            },
          }
        );
        const secrets = secretsResponse.ok ? await secretsResponse.json() : [];

        return [
          env,
          {
            variables: this.arrayToObject(variables),
            secrets: this.arrayToObject(secrets),
          },
        ];
      });

      const results = await Promise.all(promises);
      this.metas = Object.fromEntries(results);

      this.renderCompareTable();
    } catch (error) {
      this.showToast("Failed to load environment data", "error");
      console.error("Load meta error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  arrayToObject(arr) {
    const obj = {};
    if (!arr || !Array.isArray(arr)) {
      return obj;
    }
    arr.forEach((item) => {
      if (item && item.name) {
        obj[item.name] = item.value || "•••••••";
      }
    });
    return obj;
  }

  async loadRepoScopeData() {
    if (!this.ownerRepo.owner || !this.ownerRepo.name) return;

    try {
      // Load repository variables and secrets for the repo scope panel
      const variablesResponse = await fetch(
        `/api/repos/${this.ownerRepo.owner}/${this.ownerRepo.name}/variables`,
        {
          headers: {
            "X-Session-ID": this.sessionId || "",
          },
        }
      );
      const variables = variablesResponse.ok
        ? await variablesResponse.json()
        : [];

      const secretsResponse = await fetch(
        `/api/repos/${this.ownerRepo.owner}/${this.ownerRepo.name}/secrets`,
        {
          headers: {
            "X-Session-ID": this.sessionId || "",
          },
        }
      );
      const secrets = secretsResponse.ok ? await secretsResponse.json() : [];

      this.renderRepoTable(variables, secrets);
    } catch (error) {
      console.error("Load repo scope data error:", error);
      this.showToast(
        "Failed to load repository variables and secrets",
        "error"
      );
    }
  }

  renderRepoTable(variables, secrets) {
    const repoTable = document.getElementById("repoTable");
    if (!repoTable) return;

    // Ensure variables and secrets are arrays
    const vars = Array.isArray(variables) ? variables : [];
    const secs = Array.isArray(secrets) ? secrets : [];

    if (vars.length === 0 && secs.length === 0) {
      repoTable.innerHTML = `
        <div class="p-4 text-center text-neutral-500">
          <div class="text-sm">No variables or secrets found</div>
          <div class="text-xs mt-1">Add some using the form above</div>
        </div>
      `;
      return;
    }

    let html = '<div class="space-y-4">';

    // Variables Section
    if (vars.length > 0) {
      html += `
        <div class="border rounded-lg">
          <div class="bg-blue-50 border-b px-4 py-2">
            <h4 class="text-sm font-medium text-blue-900">🔧 Variables (${vars.length})</h4>
          </div>
          <div class="overflow-x-auto max-h-96 overflow-y-auto">
            <table class="w-full text-left">
              <thead class="bg-gray-50 text-xs sticky top-0">
                <tr>
                  <th class="px-3 py-2">Name</th>
                  <th class="px-3 py-2">Value</th>
                  <th class="px-3 py-2">Updated</th>
                  <th class="px-3 py-2 text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
      `;

      vars.forEach((variable) => {
        html += `
          <tr class="border-t">
            <td class="px-3 py-2 text-xs font-mono">${variable.name}</td>
            <td class="px-3 py-2 text-xs font-mono flex items-center gap-2">
              <span class="flex-1">${variable.value}</span>
              <button class="p-1 text-slate-400 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors" 
                      onclick="app.copyToClipboard('${variable.value.replace(
                        /'/g,
                        "\\'"
                      )}')" 
                      title="Copy value">
                <i class="fas fa-copy text-xs"></i>
              </button>
              <button class="p-1 text-slate-400 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors" 
                      onclick="app.editRepoItem('${
                        variable.name
                      }', 'variable', '${variable.value.replace(
          /'/g,
          "\\'"
        )}')" 
                      title="Edit variable">
                <i class="fas fa-edit text-xs"></i>
              </button>
            </td>
            <td class="px-3 py-2 text-xs">${new Date(
              variable.updated_at
            ).toLocaleDateString()}</td>
            <td class="px-3 py-2 text-right">
              <button class="px-2 py-1 rounded text-xs bg-red-100 text-red-700 hover:bg-red-200" 
                      onclick="app.deleteRepoItem('${
                        variable.name
                      }', 'variable')">
                Delete
              </button>
            </td>
          </tr>
        `;
      });

      html += `
              </tbody>
            </table>
          </div>
        </div>
      `;
    }

    // Secrets Section
    if (secs.length > 0) {
      html += `
        <div class="border rounded-lg">
          <div class="bg-orange-50 border-b px-4 py-2">
            <h4 class="text-sm font-medium text-orange-900">🔒 Secrets (${secs.length})</h4>
          </div>
          <div class="overflow-x-auto max-h-96 overflow-y-auto">
            <table class="w-full text-left">
              <thead class="bg-gray-50 text-xs sticky top-0">
                <tr>
                  <th class="px-3 py-2">Name</th>
                  <th class="px-3 py-2">Value</th>
                  <th class="px-3 py-2">Updated</th>
                  <th class="px-3 py-2 text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
      `;

      secs.forEach((secret) => {
        html += `
          <tr class="border-t">
            <td class="px-3 py-2 text-xs font-mono">${secret.name}</td>
            <td class="px-3 py-2 text-xs font-mono text-neutral-400 flex items-center gap-2">
              <span class="flex-1">••••••••</span>
              <button class="p-1 text-slate-400 hover:text-orange-600 hover:bg-orange-50 rounded transition-colors" 
                      onclick="app.editRepoItem('${secret.name}', 'secret')" 
                      title="Edit secret">
                <i class="fas fa-edit text-xs"></i>
              </button>
            </td>
            <td class="px-3 py-2 text-xs">${new Date(
              secret.updated_at
            ).toLocaleDateString()}</td>
            <td class="px-3 py-2 text-right">
              <button class="px-2 py-1 rounded text-xs bg-red-100 text-red-700 hover:bg-red-200" 
                      onclick="app.deleteRepoItem('${secret.name}', 'secret')">
                Delete
              </button>
            </td>
          </tr>
        `;
      });

      html += `
              </tbody>
            </table>
          </div>
        </div>
      `;
    }

    html += "</div>";
    repoTable.innerHTML = html;
  }

  renderCompareTable() {
    const compareTable = document.getElementById("compareTable");

    if (this.selectedEnvs.length === 0) {
      compareTable.innerHTML = `
        <div class="empty-state">
          <i class="fas fa-layer-group"></i>
          <h3>No environments selected</h3>
          <p>Select environments above to compare variables and secrets across your deployment environments.</p>
          <button onclick="app.selectAllEnvironments()" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
            Select All Environments
          </button>
        </div>
      `;
      return;
    }

    // Calculate counts for each environment
    const envCounts = {};
    this.selectedEnvs.forEach((env) => {
      const meta = this.metas[env] || { variables: {}, secrets: {} };
      envCounts[env] = {
        variables: Object.keys(meta.variables).length,
        secrets: Object.keys(meta.secrets).length,
      };
    });

    const keys = this.unionKeys();
    const variableKeys = keys.filter((key) =>
      Object.values(this.metas).some((meta) => key in meta.variables)
    );
    const secretKeys = keys.filter((key) =>
      Object.values(this.metas).some((meta) => key in meta.secrets)
    );

    let html = '<div class="space-y-6">';

    // Variables Section
    if (variableKeys.length > 0) {
      html += `
        <div class="card">
          <div class="bg-gradient-to-r from-blue-50 to-indigo-50 border-b border-blue-200 px-6 py-4">
            <h4 class="text-base font-semibold text-blue-900 flex items-center gap-2">
              <i class="fas fa-wrench text-blue-600"></i>
              Variables (${variableKeys.length})
            </h4>
            <p class="text-xs text-blue-700 mt-1">Environment variables for your applications</p>
          </div>
          <div class="table-container overflow-x-auto max-h-96 overflow-y-auto">
            <table class="w-full">
              <thead class="sticky top-0 bg-white">
                <tr>
                  <th class="text-left">Key</th>
                  ${this.selectedEnvs
                    .map(
                      (env) =>
                        `<th class="text-left">${env} <span class="text-blue-600 font-normal">(${envCounts[env].variables})</span></th>`
                    )
                    .join("")}
                  <th class="text-left">Actions</th>
                </tr>
              </thead>
              <tbody>
      `;

      variableKeys.forEach((key) => {
        html += `<tr class="table-row-enter">`;
        html += `<td class="font-mono text-sm font-medium text-slate-800">${key}</td>`;

        this.selectedEnvs.forEach((env) => {
          const meta = this.metas[env] || { variables: {}, secrets: {} };
          const hasVariable = key in meta.variables;

          html += `<td>`;
          if (hasVariable) {
            html += `<div class="flex items-center gap-2">
              <code class="px-2 py-1 bg-slate-100 text-slate-800 rounded text-xs flex-1">${
                meta.variables[key]
              }</code>
              <button class="p-1 text-slate-400 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors" 
                      onclick="app.copyToClipboard('${meta.variables[
                        key
                      ].replace(/'/g, "\\'")}')" 
                      title="Copy value">
                <i class="fas fa-copy text-xs"></i>
              </button>
              <button class="p-1 text-slate-400 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors" 
                      onclick="app.editKey('${env}', '${key}', 'variable', '${meta.variables[
              key
            ].replace(/'/g, "\\'")}')" 
                      title="Edit variable">
                <i class="fas fa-edit text-xs"></i>
              </button>
            </div>`;
          } else {
            html += `<button class="px-3 py-1 rounded-md text-xs font-medium bg-blue-50 text-blue-700 border border-blue-200 hover:bg-blue-100 transition-colors" onclick="app.addKeyHere('${env}', '${key}')">
              <i class="fas fa-plus text-xs mr-1"></i>Add here
            </button>`;
          }
          html += "</td>";
        });

        html += '<td class="min-w-32">';
        html += '<div class="flex flex-col gap-2">';
        html += `<select onchange="app.syncFrom(this.value, '${key}')" class="text-xs rounded-md border-slate-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 w-full">`;
        html += '<option value="">Sync from…</option>';
        this.selectedEnvs.forEach((env) => {
          html += `<option value="${env}">${env}</option>`;
        });
        html += "</select>";
        html += `<select onchange="app.deleteIn(this.value, '${key}')" class="text-xs rounded-md border-slate-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 w-full">`;
        html += '<option value="">Delete in…</option>';
        this.selectedEnvs.forEach((env) => {
          html += `<option value="${env}">${env}</option>`;
        });
        html += "</select>";
        html += "</div></td>";
        html += "</tr>";
      });

      html += `
              </tbody>
            </table>
          </div>
        </div>
      `;
    }

    // Secrets Section
    if (secretKeys.length > 0) {
      html += `
        <div class="card">
          <div class="bg-gradient-to-r from-orange-50 to-red-50 border-b border-orange-200 px-6 py-4">
            <h4 class="text-base font-semibold text-orange-900 flex items-center gap-2">
              <i class="fas fa-lock text-orange-600"></i>
              Secrets (${secretKeys.length})
            </h4>
            <p class="text-xs text-orange-700 mt-1">Encrypted secrets for secure deployment</p>
          </div>
          <div class="table-container overflow-x-auto max-h-96 overflow-y-auto">
            <table class="w-full">
              <thead class="sticky top-0 bg-white">
                <tr>
                  <th class="text-left">Key</th>
                  ${this.selectedEnvs
                    .map(
                      (env) =>
                        `<th class="text-left">${env} <span class="text-orange-600 font-normal">(${envCounts[env].secrets})</span></th>`
                    )
                    .join("")}
                  <th class="text-left">Actions</th>
                </tr>
              </thead>
              <tbody>
      `;

      secretKeys.forEach((key) => {
        html += `<tr class="table-row-enter">`;
        html += `<td class="font-mono text-sm font-medium text-slate-800">${key}</td>`;

        this.selectedEnvs.forEach((env) => {
          const meta = this.metas[env] || { variables: {}, secrets: {} };
          const hasSecret = key in meta.secrets;

          html += `<td>`;
          if (hasSecret) {
            html += `<div class="flex items-center gap-2">
              <span class="inline-flex items-center px-2 py-1 bg-slate-100 text-slate-600 rounded text-xs font-mono flex-1">
                <i class="fas fa-eye-slash text-xs mr-1"></i>••••••••
              </span>
              <button class="p-1 text-slate-400 hover:text-orange-600 hover:bg-orange-50 rounded transition-colors" 
                      onclick="app.editKey('${env}', '${key}', 'secret')" 
                      title="Edit secret">
                <i class="fas fa-edit text-xs"></i>
              </button>
            </div>`;
          } else {
            html += `<button class="px-3 py-1 rounded-md text-xs font-medium bg-orange-50 text-orange-700 border border-orange-200 hover:bg-orange-100 transition-colors" onclick="app.addKeyHere('${env}', '${key}')">
              <i class="fas fa-plus text-xs mr-1"></i>Add here
            </button>`;
          }
          html += "</td>";
        });

        html += '<td class="min-w-32">';
        html += '<div class="flex flex-col gap-2">';
        html += `<select onchange="app.syncFrom(this.value, '${key}')" class="text-xs rounded-md border-slate-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 w-full">`;
        html += '<option value="">Sync from…</option>';
        this.selectedEnvs.forEach((env) => {
          html += `<option value="${env}">${env}</option>`;
        });
        html += "</select>";
        html += `<select onchange="app.deleteIn(this.value, '${key}')" class="text-xs rounded-md border-slate-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 w-full">`;
        html += '<option value="">Delete in…</option>';
        this.selectedEnvs.forEach((env) => {
          html += `<option value="${env}">${env}</option>`;
        });
        html += "</select>";
        html += "</div></td>";
        html += "</tr>";
      });

      html += `
              </tbody>
            </table>
          </div>
        </div>
      `;
    }

    html += "</div>";
    compareTable.innerHTML = html;
  }

  unionKeys() {
    const keys = new Set();
    Object.values(this.metas).forEach((meta) => {
      Object.keys(meta.variables).forEach((key) => keys.add(key));
      Object.keys(meta.secrets).forEach((key) => keys.add(key));
    });

    // Sort keys to show variables first, then secrets
    const allKeys = Array.from(keys);
    const variableKeys = allKeys.filter((key) => {
      // Check if this key exists as a variable in any environment
      return Object.values(this.metas).some((meta) => key in meta.variables);
    });
    const secretKeys = allKeys.filter((key) => {
      // Check if this key exists as a secret in any environment (and not as a variable)
      return (
        !variableKeys.includes(key) &&
        Object.values(this.metas).some((meta) => key in meta.secrets)
      );
    });

    return [...variableKeys.sort(), ...secretKeys.sort()];
  }

  async handleAddKey(e) {
    e.preventDefault();

    const key = document.getElementById("newKey").value.trim();
    const type = document.getElementById("newKeyType").value;
    const value = document.getElementById("newKeyValue").value.trim();
    const targets = this.targetEnvs;

    if (!key || !value || !targets.length) {
      this.showToast(
        "Key, value, and at least one environment are required",
        "error"
      );
      return;
    }

    this.showLoading(true);
    try {
      for (const env of targets) {
        await this.upsertKey({ env, key, value, type });
      }
      await this.loadMeta();
      this.showToast(`Created ${key} in ${targets.join(", ")}`, "success");

      // Clear form
      document.getElementById("newKey").value = "";
      document.getElementById("newKeyValue").value = "";
      // Clear target environments selection
      this.targetEnvs = [];
      this.renderTargetEnvs();
    } catch (error) {
      this.showToast("Failed to create key", "error");
      console.error("Add key error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  async upsertKey({ env, key, value, type }) {
    let method, url;

    if (env === null) {
      // Repository-wide operation
      method = "POST";
      url = `/api/repos/${this.ownerRepo.owner}/${this.ownerRepo.name}/${
        type === "secret" ? "secrets" : "variables"
      }`;
    } else {
      // Environment-specific operation
      const existingMeta = this.metas && this.metas[env];
      const exists =
        existingMeta &&
        (existingMeta.variables[key] || existingMeta.secrets[key]);

      method = exists ? "PUT" : "POST";
      url = exists
        ? `/api/repos/${this.ownerRepo.owner}/${
            this.ownerRepo.name
          }/environments/${env}/${
            type === "secret" ? "secrets" : "variables"
          }/${encodeURIComponent(key)}`
        : `/api/repos/${this.ownerRepo.owner}/${
            this.ownerRepo.name
          }/environments/${env}/${type === "secret" ? "secrets" : "variables"}`;
    }

    const body = method === "PUT" ? { value } : { name: key, value };

    const response = await fetch(url, {
      method,
      headers: {
        "Content-Type": "application/json",
        "X-Session-ID": this.sessionId || "",
      },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const target = env === null ? "repository" : env;
      throw new Error(
        `Failed to ${
          method === "PUT" ? "update" : "create"
        } ${type} in ${target}`
      );
    }
  }

  async copyToClipboard(text) {
    try {
      await navigator.clipboard.writeText(text);
      this.showToast("Value copied to clipboard", "success");
    } catch (err) {
      // Fallback for older browsers
      const textArea = document.createElement("textarea");
      textArea.value = text;
      document.body.appendChild(textArea);
      textArea.select();
      try {
        document.execCommand("copy");
        this.showToast("Value copied to clipboard", "success");
      } catch (fallbackErr) {
        this.showToast("Failed to copy value", "error");
      }
      document.body.removeChild(textArea);
    }
  }

  async addKeyHere(env, key) {
    const value = await this.showPrompt(
      "Add Variable/Secret",
      `Enter value for <strong>${key}</strong> in environment <strong>${env}</strong>:`,
      ""
    );
    if (value === null) return;

    const type = await this.showPrompt(
      "Select Type",
      "Choose the type for this key:",
      "variable"
    );
    if (!type) return;

    this.showLoading(true);
    try {
      await this.upsertKey({ env, key, value, type });
      await this.loadMeta();
      this.showToast(`Added ${key} to ${env}`, "success");
    } catch (error) {
      this.showToast("Failed to add key", "error");
      console.error("Add key here error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  async syncFrom(fromEnv, key) {
    if (!fromEnv) return;

    const meta = this.metas[fromEnv];
    if (!meta) return;

    const type =
      key in meta.variables
        ? "variable"
        : key in meta.secrets
        ? "secret"
        : null;
    if (!type) {
      this.showToast(`Key ${key} not found in ${fromEnv}`, "error");
      return;
    }

    const value =
      type === "variable"
        ? meta.variables[key]
        : await this.showPrompt(
            "Sync Secret",
            `Enter value for secret <strong>${key}</strong> to sync from <strong>${fromEnv}</strong>:`,
            ""
          );
    if (value === null) return;

    const targets = this.selectedEnvs.filter((env) => {
      const m = this.metas[env] || { variables: {}, secrets: {} };
      return !(key in m.variables) && !(key in m.secrets);
    });

    this.showLoading(true);
    try {
      for (const env of targets) {
        await this.upsertKey({ env, key, value, type });
      }
      await this.loadMeta();
      this.showToast(`Synced ${key} to ${targets.join(", ")}`, "success");
    } catch (error) {
      this.showToast("Failed to sync key", "error");
      console.error("Sync error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  async deleteIn(env, key) {
    if (!env) return;

    const confirmed = await this.showConfirm(
      "Delete Variable/Secret",
      `Are you sure you want to delete <strong>${key}</strong> from environment <strong>${env}</strong>?`
    );
    if (!confirmed) return;

    const meta = this.metas[env] || { variables: {}, secrets: {} };
    const type = key in meta.variables ? "variable" : "secret";

    this.showLoading(true);
    try {
      const response = await fetch(
        `/api/repos/${this.ownerRepo.owner}/${
          this.ownerRepo.name
        }/environments/${env}/${
          type === "secret" ? "secrets" : "variables"
        }/${encodeURIComponent(key)}`,
        {
          method: "DELETE",
          headers: {
            "X-Session-ID": this.sessionId || "",
          },
        }
      );

      if (response.ok) {
        await this.loadMeta();
        this.showToast(`Deleted ${key} from ${env}`, "success");
      } else {
        this.showToast("Failed to delete key", "error");
      }
    } catch (error) {
      this.showToast("Failed to delete key", "error");
      console.error("Delete error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  async editKey(env, key, type, currentValue = null) {
    let value;

    if (type === "variable") {
      value = await this.showPrompt(
        "✏️ Edit Variable",
        `Key: <strong>${key}</strong><br>Environment: <strong>${env}</strong><br><br>Enter new value:`,
        currentValue || ""
      );
    } else {
      value = await this.showPrompt(
        "🔒 Edit Secret",
        `Key: <strong>${key}</strong><br>Environment: <strong>${env}</strong><br><br>Enter new secret value:`,
        ""
      );
    }

    if (value === null) return;

    this.showLoading(true);
    try {
      await this.upsertKey({ env, key, value, type });
      await this.loadMeta();
      this.showToast(`✅ Updated ${key} in ${env}`, "success");
    } catch (error) {
      this.showToast("❌ Failed to update key", "error");
      console.error("Edit key error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  async editRepoItem(name, type, currentValue = null) {
    let value;

    if (type === "variable") {
      value = await this.showPrompt(
        "✏️ Edit Repository Variable",
        `Name: <strong>${name}</strong><br><br>Enter new value:`,
        currentValue || ""
      );
    } else {
      value = await this.showPrompt(
        "🔒 Edit Repository Secret",
        `Name: <strong>${name}</strong><br><br>Enter new secret value:`,
        ""
      );
    }

    if (value === null) return;

    this.showLoading(true);
    try {
      const response = await fetch(
        `/api/repos/${this.ownerRepo.owner}/${this.ownerRepo.name}/${
          type === "secret" ? "secrets" : "variables"
        }/${encodeURIComponent(name)}`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            "X-Session-ID": this.sessionId || "",
          },
          body: JSON.stringify({ value }),
        }
      );

      if (response.ok) {
        await this.loadRepoScopeData();
        this.showToast(`✅ Updated ${name}`, "success");
      } else {
        this.showToast("❌ Failed to update item", "error");
      }
    } catch (error) {
      this.showToast("❌ Failed to update item", "error");
      console.error("Edit repo item error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  async exportEnvironment(env) {
    this.exporting = true;
    this.showLoading(true);

    try {
      // Get the environment data from our loaded metas
      const meta = this.metas[env];
      if (!meta) {
        this.showToast(`No data found for environment: ${env}`, "error");
        return;
      }

      // Combine variables and secrets (only variables for .env export)
      const envContent = Object.entries(meta.variables)
        .map(([key, value]) => `${key}=${value}`)
        .join("\n");

      this.exportText = `# ${this.ownerRepo.owner}/${
        this.ownerRepo.name
      } :: ${env}\n# Exported on ${new Date().toISOString()}\n\n${envContent}`;
      this.showExportPreview();

      // Download file
      const blob = new Blob([this.exportText], { type: "text/plain" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `${this.ownerRepo.name}-${env}.env`;
      a.click();
      URL.revokeObjectURL(url);

      this.showToast(
        `Exported ${env}.env with ${
          Object.keys(meta.variables).length
        } variables`,
        "success"
      );
    } catch (error) {
      this.showToast("Failed to export environment", "error");
      console.error("Export error:", error);
    } finally {
      this.exporting = false;
      this.showLoading(false);
    }
  }

  showExportPreview() {
    document.getElementById("exportText").value = this.exportText;
    document.getElementById("exportPreview").classList.remove("hidden");
  }

  handleImportFile(e) {
    const file = e.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      this.parseDotEnv(String(e.target?.result || ""));
    };
    reader.readAsText(file);
  }

  handleImportText(e) {
    this.parseDotEnv(e.target.value);
  }

  parseDotEnv(text) {
    const result = {};
    const lines = (text || "").split(/\r?\n/);

    for (let raw of lines) {
      let line = raw.trim();
      if (!line || line.startsWith("#")) continue;

      const idx = line.indexOf("=");
      if (idx === -1) continue;

      const key = line.slice(0, idx).trim();
      let val = line.slice(idx + 1);
      if (
        (val.startsWith('"') && val.endsWith('"')) ||
        (val.startsWith("'") && val.endsWith("'"))
      ) {
        val = val.slice(1, -1);
      }
      result[key] = val;
    }

    this.importPreview = result;
    this.showImportPreview();
  }

  showImportPreview() {
    const preview = document.getElementById("importPreview");
    const content = document.getElementById("importPreviewContent");
    const count = document.getElementById("importCount");

    if (Object.keys(this.importPreview).length > 0) {
      count.textContent = Object.keys(this.importPreview).length;
      content.innerHTML = Object.entries(this.importPreview)
        .map(
          ([k, v]) =>
            `<div class="flex items-center justify-between gap-2"><span>${k}</span><span class="truncate">= ${String(
              v
            )}</span></div>`
        )
        .join("");
      preview.classList.remove("hidden");
    } else {
      preview.classList.add("hidden");
    }
  }

  async applyImport() {
    if (!Object.keys(this.importPreview).length) {
      this.showToast("Nothing to import", "error");
      return;
    }

    if (!this.importTargets.length) {
      this.showToast("Select at least one environment", "error");
      return;
    }

    this.showLoading(true);
    try {
      for (const env of this.importTargets) {
        for (const [key, value] of Object.entries(this.importPreview)) {
          await this.upsertKey({ env, key, value, type: "variable" });
        }
      }
      await this.loadMeta();
      this.showToast(
        `Imported ${
          Object.keys(this.importPreview).length
        } vars into ${this.importTargets.join(", ")}`,
        "success"
      );
      this.clearImport();
    } catch (error) {
      this.showToast("Failed to import variables", "error");
      console.error("Import error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  clearImport() {
    this.importPreview = {};
    document.getElementById("importText").value = "";
    document.getElementById("importFile").value = "";
    this.showImportPreview();
  }

  async addRepoKey() {
    const key = document.getElementById("repoNewKey").value.trim();
    const type = document.getElementById("repoNewType").value;
    const value = document.getElementById("repoNewValue").value.trim();

    if (!key || !value) {
      this.showToast("Key and value are required", "error");
      return;
    }

    this.showLoading(true);
    try {
      await this.upsertKey({ env: null, key, value, type });
      this.showToast(`Added ${key} to repository`, "success");

      // Clear form
      document.getElementById("repoNewKey").value = "";
      document.getElementById("repoNewValue").value = "";

      // Refresh data
      await this.loadRepoScopeData();
      await this.loadMeta();
    } catch (error) {
      this.showToast("Failed to add repository key", "error");
      console.error("Add repo key error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  async deleteRepoItem(name, type) {
    const confirmed = await this.showConfirm(
      "Delete Repository Item",
      `Are you sure you want to delete <strong>${name}</strong> (${type})?`
    );
    if (!confirmed) return;

    this.showLoading(true);
    try {
      const response = await fetch(
        `/api/repos/${this.ownerRepo.owner}/${this.ownerRepo.name}/${
          type === "secret" ? "secrets" : "variables"
        }/${encodeURIComponent(name)}`,
        {
          method: "DELETE",
          headers: {
            "X-Session-ID": this.sessionId || "",
          },
        }
      );

      if (!response.ok) {
        throw new Error(`Failed to delete ${type}`);
      }

      await this.loadRepoScopeData();
      await this.loadMeta(); // Refresh environment view too
      this.showToast(`Deleted ${name}`, "success");
    } catch (error) {
      this.showToast("Failed to delete item", "error");
      console.error("Delete repo item error:", error);
    } finally {
      this.showLoading(false);
    }
  }

  async refreshMeta() {
    await this.loadMeta();
  }

  showLoading(show) {
    const spinner = document.getElementById("loadingSpinner");
    if (show) {
      spinner.classList.remove("hidden");
    } else {
      spinner.classList.add("hidden");
    }
  }

  showToast(message, type = "info") {
    const container = document.getElementById("toastContainer");
    const toast = document.createElement("div");
    toast.className = `toast ${type}`;
    toast.innerHTML = `
      <div class="flex items-center">
        <i class="fas fa-${this.getToastIcon(type)} mr-2"></i>
        <span>${message}</span>
      </div>
    `;

    container.appendChild(toast);

    setTimeout(() => {
      toast.remove();
    }, 5000);
  }

  // Modern prompt system
  showPrompt(title, content, defaultValue = "") {
    return new Promise((resolve, reject) => {
      const modal = document.getElementById("promptModal");
      const titleEl = document.getElementById("promptTitle");
      const contentEl = document.getElementById("promptContent");
      const confirmBtn = document.getElementById("promptConfirmBtn");

      titleEl.textContent = title;

      // Create input field
      contentEl.innerHTML = `
        <div class="space-y-3">
          <div class="text-sm text-gray-600">
            ${content}
          </div>
          <input type="text" id="promptInput" 
                 class="w-full px-4 py-3 rounded-xl border border-slate-300 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 transition-all font-mono text-sm"
                 placeholder="Enter value..."
                 value="${defaultValue}">
        </div>
      `;

      modal.classList.remove("hidden");

      // Focus input
      setTimeout(() => {
        const input = document.getElementById("promptInput");
        input.focus();
        input.select();
      }, 100);

      // Handle confirm
      const handleConfirm = () => {
        const input = document.getElementById("promptInput");
        const value = input.value.trim();
        modal.classList.add("hidden");
        resolve(value);
      };

      // Handle cancel
      const handleCancel = () => {
        modal.classList.add("hidden");
        resolve(null);
      };

      // Event listeners
      confirmBtn.onclick = handleConfirm;

      // Enter key to confirm, Escape to cancel
      const input = document.getElementById("promptInput");
      input.onkeydown = (e) => {
        if (e.key === "Enter") {
          handleConfirm();
        } else if (e.key === "Escape") {
          handleCancel();
        }
      };

      // Close button
      modal.querySelector('[onclick="app.closePromptModal()"]').onclick =
        handleCancel;
    });
  }

  closePromptModal() {
    const modal = document.getElementById("promptModal");
    modal.classList.add("hidden");
  }

  // Modern confirm system
  showConfirm(title, message) {
    return new Promise((resolve) => {
      const modal = document.getElementById("promptModal");
      const titleEl = document.getElementById("promptTitle");
      const contentEl = document.getElementById("promptContent");
      const confirmBtn = document.getElementById("promptConfirmBtn");

      titleEl.textContent = title;
      contentEl.innerHTML = `
        <div class="text-sm text-gray-600">
          ${message}
        </div>
      `;

      confirmBtn.textContent = "Confirm";
      confirmBtn.className =
        "px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-xl transition-colors font-medium";

      modal.classList.remove("hidden");

      const handleConfirm = () => {
        modal.classList.add("hidden");
        resolve(true);
      };

      const handleCancel = () => {
        modal.classList.add("hidden");
        resolve(false);
      };

      confirmBtn.onclick = handleConfirm;
      modal.querySelector('[onclick="app.closePromptModal()"]').onclick =
        handleCancel;
    });
  }

  getToastIcon(type) {
    switch (type) {
      case "success":
        return "check-circle";
      case "error":
        return "exclamation-circle";
      case "warning":
        return "exclamation-triangle";
      default:
        return "info-circle";
    }
  }
}

// Global functions for modal closing
function closeTokenModal() {
  app.closeTokenModal();
}

function closeCreateEnvModal() {
  app.closeCreateEnvModal();
}

// Initialize the application
const app = new GitHubEnvManager();
