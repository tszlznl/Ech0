<p align="left">
  <a href="https://hellogithub.com/repository/lin-snow/Ech0" target="_blank">
    <img src="https://api.hellogithub.com/v1/widgets/recommend.svg?rid=8f3cafdd6ef3445dbb1c0ed6dd34c8b5&claim_uid=swhbQfnJvKS0t7I&theme=neutral"
         alt="Featured｜HelloGitHub"
         width="250"
         height="54" />
  </a>
</p>

<p align="right">
  <a title="zh" href="./README.md">
    <img src="https://img.shields.io/badge/-简体中文-545759?style=for-the-badge" alt="简体中文">
  </a>
  <img src="https://img.shields.io/badge/-English-F54A00?style=for-the-badge" alt="English">
</p>



<div align="center">
  <img alt="Ech0" src="./docs/imgs/logo.svg" width="150">

  [Preview](https://memo.vaaat.com/) | [Official Site & Doc](https://www.ech0.app/) | [Ech0 Hub](https://hub.ech0.app/)

  # Ech0
</div>

<div align="center">

[![GitHub release](https://img.shields.io/github/v/release/lin-snow/Ech0)](https://github.com/lin-snow/Ech0/releases) ![License](https://img.shields.io/github/license/lin-snow/Ech0) [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/lin-snow/Ech0) [![Hello Github](https://api.hellogithub.com/v1/widgets/recommend.svg?rid=8f3cafdd6ef3445dbb1c0ed6dd34c8b5&claim_uid=swhbQfnJvKS0t7I&theme=small)](https://hellogithub.com/repository/lin-snow/Ech0)

</div>

> A next-generation open-source, self-hosted, lightweight publishing platform focused on personal idea sharing.

Ech0 is a new-generation open-source self-hosted platform designed for individual users. It is ultra-lightweight and low-cost, helping you easily publish and share ideas, writings, and links. With a clean, intuitive interface and powerful command-line tools, content management becomes simple and flexible. Your data is fully owned and controlled by you.

![Interface Preview](./docs/imgs/screenshot.png)

---

<details>
   <summary><strong>Table of Contents</strong></summary>

- [Ech0](#ech0)
  - [Highlights](#highlights)
  - [Quick Deployment](#quick-deployment)
    - [🐳 Docker (Recommended)](#-docker-recommended)
    - [🐋 Docker Compose](#-docker-compose)
    - [☸️ Kubernetes (Helm)](#️-kubernetes-helm)
  - [Upgrading](#upgrading)
    - [🔄 Docker](#-docker)
    - [💎 Docker Compose](#-docker-compose-1)
    - [☸️ Kubernetes (Helm)](#️-kubernetes-helm-1)
  - [FAQ](#faq)
  - [Feedback \& Community](#feedback--community)
  - [Architecture](#architecture)
  - [Development Guide](#development-guide)
    - [Backend Requirements](#backend-requirements)
    - [Frontend Requirements](#frontend-requirements)
    - [Start Backend \& Frontend](#start-backend--frontend)
  - [Thanks for Your Support!](#thanks-for-your-support)
  - [Star History](#star-history)
  - [Acknowledgements](#acknowledgements)
  - [Support](#support)
</details>

---

## Highlights

- ☁️ **Lightweight, Efficient Architecture**: Low resource usage and compact images fit environments from personal servers to ARM devices.  
- 🚀 **Fast Deployment Experience**: Docker-first, out-of-the-box deployment from install to run in a single command.  
- 📦 **Self-Contained Distribution**: Complete binaries and container images run without extra runtime dependencies.  
- 💻 **Cross-Platform Support**: Supports Linux, Windows, and ARM devices (for example, Raspberry Pi).  

## Storage & Data

- 🗂️ **VireFS Unified Storage Abstraction**: **VireFS** unifies mounting and management across local storage and S3-compatible object storage.  
- ☁️ **S3 Object Storage Support**: Native support for S3-compatible object storage for cloud-scale asset expansion.  
- 📦 **Data Sovereignty Architecture**: All content and metadata remain user-owned, with RSS output support.  
- 🔄 **Data Migration Workflow**: Import historical data through migration flows and pair with snapshot export for archiving.  
- 🔐 **Automated Backup System**: Full export and backup via Web, CLI, and TUI, plus automatic background backups.  

## Writing & Content

- ✍️ **Markdown Authoring Experience**: A **markdown-it**-based editor and renderer with plugin extensibility and live preview.  
- 🧘 **Zen Mode Immersive Reading**: A low-distraction Timeline reading mode designed for focused browsing.  
- 🏷️ **Tag Management System**: Tag-based organization with fast filtering and precise content retrieval.  
- 🃏 **Rich Media Cards**: Card-based presentation for links, GitHub projects, and other rich content.  
- 🎥 **Video Content Parsing**: Built-in parsing and display for Bilibili and YouTube content.  

## Media & Assets

- 📁 **Visual File Manager**: Built-in manager for file upload, browsing, and media asset handling.  

## Social & Interaction

- 💬 **Native Comment System**: Built-in comments and moderation without relying on third-party comment services.  
- 🃏 **Social Interaction Features**: Supports interactions such as likes and sharing.  

## Auth & Security

- 🔑 **OAuth2 / OIDC Authentication**: Supports OAuth2 and OIDC for integration with third-party identity providers.  
- 🙈 **Passkey Passwordless Login**: Supports passkey authentication via biometrics or hardware security keys.  
- 🔑 **Access Token Management**: Generate and revoke access tokens for API calls and third-party integrations.  
- 👤 **Multi-Account Permissions**: Supports multi-user management with permission control.  

## System & Developer

- 🧱 **Busen Data Bus Architecture**: A self-built Busen data bus enables decoupled modules and reliable message delivery.  
- 📊 **Structured Logging System**: Standardized structured logs improve readability, observability, and analysis.  
- 🖥️ **Real-Time Log Console**: Built-in web console for live log streaming, debugging, and incident diagnosis.  
- 📟 **TUI Management Interface**: Terminal UI for convenient server-side administration.  
- 🧰 **CLI Toolchain**: CLI utilities for automation workflows and script integration.  
- 🔗 **Open API & Webhook**: Complete API and Webhook support for system integrations and automated workflows.  

## Experience

- 🌍 **Cross-Device Adaptation**: Responsive design for desktop, tablet, and mobile browsers.  
- 👾 **PWA Support**: Install as a web app for a more native-like experience.  
- 🌗 **Themes & Dark Mode**: Supports dark mode and extensible theming.  

## License

- 🎉 **Fully Open Source**: Released under **AGPL-3.0**, with no tracking, no subscription, and no SaaS dependency.  


---

## Quick Deployment
<!-- 
### 🧙 One-Click Script Deployment (Recommended, make sure your network can access GitHub Release)
```shell
curl -fsSL "https://sh.soopy.cn/ech0.sh" -o ech0.sh && bash ech0.sh
``` -->

### 🐳 Docker (Recommended)

```shell
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

> 💡 After deployment, access `ip:6277` to use  
> 🚷 It is recommended to change `JWT_SECRET="Hello Echos"` to a secure secret  
> 📍 The first registered user will be set as administrator  
> 🎈 Data stored under `/opt/ech0/data`

### 🐋 Docker Compose

1. Create a new directory and place `docker-compose.yml` inside.  
2. Run:

```shell
docker-compose up -d
```

### ☸️ Kubernetes (Helm)

If you want to deploy Ech0 in a Kubernetes cluster, you can use the Helm Chart provided in this project.

Since this project does not provide an online Helm repository, you need to clone the repository to your local machine first, and then install from the local directory.

1.  **Clone the repository:**
    ```shell
    git clone https://github.com/lin-snow/Ech0.git
    cd Ech0
    ```

2.  **Install with Helm:**
    ```shell
    # helm install <release-name> <chart-directory>
    helm install ech0 ./charts/ech0
    ```

    You can also customize the release name and namespace:
    ```shell
    helm install my-ech0 ./charts/ech0 --namespace my-namespace --create-namespace
    ```

---

## Upgrading

### 🔄 Docker

```shell
docker stop ech0
docker rm ech0
docker pull sn0wl1n/ech0:latest
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

### 💎 Docker Compose

```shell
cd /path/to/compose
docker-compose pull && \
docker-compose up -d --force-recreate
docker image prune -f
```

### ☸️ Kubernetes (Helm)

1.  **Update the repository:**
    Navigate to your local Ech0 repository directory and pull the latest changes.
    ```shell
    cd Ech0
    git pull
    ```

2.  **Upgrade the Helm Release:**
    Use the `helm upgrade` command to update your release.
    ```shell
    # helm upgrade <release-name> <chart-directory>
    helm upgrade ech0 ./charts/ech0
    ```
    If you used a custom release name and namespace, use the corresponding names:
    ```shell
    helm upgrade my-ech0 ./charts/ech0 --namespace my-namespace
    ```

<!-- ---

## Access Modes

### 🖥️ TUI Mode

![TUI Mode](./docs/imgs/tui.png)

Run the binary directly (for example, on Windows double-click `Ech0.exe`).

-->

---

## FAQ

1. **What is Ech0?**  
   A lightweight, open-source self-hosted platform for quickly sharing thoughts, writings, and links. All content is locally stored.  

2. **What Ech0 is NOT?**  
   Not a professional note-taking app like Obsidian or Notion; its core function is similar to social feed/microblog.  

3. **Is Ech0 free?**  
   Yes, fully free and open-source under AGPL-3.0, no ads, tracking, subscription, or service dependency.  

4. **How do I back up and import data?**  
  Since all content is stored in a local SQLite file, it is recommended to regularly back up files in `/opt/ech0/data` (or your mapped data path). In "Data Management", you can use snapshot export for archival and migration import as the only supported online data import path.

5. **Does Ech0 support RSS?**  
   Yes, content updates can be subscribed via RSS.  

6. **Why can't I publish content?**  
   Only administrators can publish. First registered user is admin.  

7. **Why no detailed permission system?**  
   Ech0 emphasizes simplicity: admin vs non-admin only, for smooth experience.  

8. **How do I set the public service URL correctly?**  
   Set your full external URL in `System Settings - Service URL` (including `http://` or `https://`) so callbacks and external integrations work correctly.  

9. **What storage backends are supported?**  
   Ech0 supports local storage by default, and can also mount S3-compatible object storage through VireFS for media and asset management.  

10. **How does data migration work in v4?**  
    Use snapshot export for archiving and use the migration flow to import historical data. Direct in-place upgrade from v3 to v4 is not supported.  

11. **What content is not recommended?**  
    Avoid publishing dense content mixing text + images + extension cards. Long posts or extension cards alone are okay.  

12. **How to enable comments?**  
    Comments are built in. Enable and configure comment-related options in system settings; no third-party comment service is required.  

13. **How to configure S3?**  
    Fill in endpoint (without http/https) and bucket with public access.

14. **How to enable passkey login?**  
  Go to `SSO - Passkey`, configure `WebAuthn RP ID` and `WebAuthn Origins`, save until status shows "Passkey ready", then bind your biometric or hardware security key following browser prompts.

---

## Feedback & Community

- Report bugs via [Issues](https://github.com/lin-snow/Ech0/issues).
- Propose features or share ideas in [Discussions](https://github.com/lin-snow/Ech0/discussions).

---

## Architecture

![Architecture Diagram](./docs/imgs/Ech0技术架构图.svg)  
> by ExcaliDraw

- The backend event bus now uses [Busen](https://github.com/lin-snow/Busen), adopting a typed-first in-process model with explicit backpressure, hooks, and drain-style shutdown.

---

## Development Guide

### Backend Requirements
- Go 1.26.0+  
- C Compiler for CGO (`go-sqlite3`):
  - Windows: [MinGW-w64](https://winlibs.com/)  
  - macOS: `brew install gcc`  
  - Linux: `sudo apt install build-essential`  
- Google Wire: `go install github.com/google/wire/cmd/wire@latest`  
- Golangci-Lint: `golangci-lint run` / `golangci-lint fmt`  
- Air (optional, backend hot reload): `make air-install` or `go install github.com/air-verse/air@latest`  
- Swagger: `swag init -g internal/server/server.go -o internal/swagger`  
- Event runtime tuning (Busen):
  - `ECH0_EVENT_DEFAULT_BUFFER` / `ECH0_EVENT_DEFAULT_OVERFLOW`
  - `ECH0_EVENT_DEADLETTER_BUFFER` / `ECH0_EVENT_SYSTEM_BUFFER`
  - `ECH0_EVENT_AGENT_BUFFER` / `ECH0_EVENT_AGENT_PARALLELISM`
  - `ECH0_EVENT_INBOX_BUFFER`
  - `ECH0_EVENT_WEBHOOK_POOL_WORKERS` / `ECH0_EVENT_WEBHOOK_POOL_QUEUE`

### Frontend Requirements
- NodeJS v25.5.0+, PNPM v10.30.0+  
- Use [fnm](https://github.com/Schniz/fnm) if multiple Node versions needed

### Start Backend & Frontend
```shell
# Backend
make run # normal backend start (equivalent to go run main.go serve)
make dev # backend hot reload with Air

# Frontend
cd web
pnpm install
pnpm dev
# or from project root: make web-dev
```

Preview: Backend `http://localhost:6277`, Frontend `http://localhost:5173`

> When importing layered packages, prefer consistent aliases such as `xxxModel`, `xxxService`, `xxxRepository`, and so on.


---

## Thanks for Your Support!

Thank you to all the friends who have supported this project! Your contributions keep it thriving 💡✨

|                        ⚙️ User                        |   🔋 Date   | 💬 Message                                       |
| :--------------------------------------------------: | :--------: | :---------------------------------------------- |
|                  🧑‍💻 Anonymous Friend                  | 2025-5-19  | Silly programmer, buy yourself some sweet drink |
|        🧑‍💻 [@sseaan](https://github.com/sseaan)        | 2025-7-27  | Ech0 is a great thing🥳                          |
| 🧑‍💻 [@QYG2297248353](https://github.com/QYG2297248353) | 2025-10-10 | None                                            |
|    🧑‍💻 [@continue33](https://github.com/continue33)    | 2025-10-23 | Thanks for fixing R2                            |
|    🧑‍💻 [@hoochanlon](https://github.com/hoochanlon)      | 2025-10-28 | None             |
|       🧑‍💻 [@Rvn0xsy](https://github.com/Rvn0xsy)       | 2025-11-12 | Great project, I will keep following! |
|                     🧑‍💻 王贼臣                     | 2025-11-20 | Thanks www.cardopt.cn             |
|       🧑‍💻 [@ljxme](https://github.com/ljxme)    | 2025-11-30 | Doing my humble part 😋             |
|       🧑‍💻 [@he9ab2l](https://github.com/he9ab2l)    | 2025-12-23 | None            |
|       🧑‍💻 鸿运当头(windfore)    | 2026-1-6 | Thank you for creating ech0           |
|       🧑‍💻 Anonymous User    | 2026-01-23  | None           |

---

## Star History

<a href="https://www.star-history.com/#lin-snow/Ech0&Timeline">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=lin-snow/Ech0&type=Timeline&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=lin-snow/Ech0&type=Timeline" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=lin-snow/Ech0&type=Timeline" />
 </picture>
</a>


---

## Acknowledgements

- Thanks to all users for their valuable suggestions and feedback.
- Thanks to all contributors and supporters from the open-source community.


![Alt](https://repobeats.axiom.co/api/embed/d69b9177e4a121e31aaed95354ff862c928ca22d.svg "Repobeats analytics image")

---

## Support

🌟 If you like **Ech0**, please give it a Star! 🚀  
Ech0 is completely free and open-source. Support helps the project continue improving.  

|                  Platform                  | QR Code                                                |
| :----------------------------------------: | :----------------------------------------------------- |
| [**Afdian**](https://afdian.com/a/l1nsn0w) | <img src="./docs/imgs/pay.jpeg" alt="Pay" width="200"> |

---

```cpp

███████╗     ██████╗    ██╗  ██╗     ██████╗ 
██╔════╝    ██╔════╝    ██║  ██║    ██╔═████╗
█████╗      ██║         ███████║    ██║██╔██║
██╔══╝      ██║         ██╔══██║    ████╔╝██║
███████╗    ╚██████╗    ██║  ██║    ╚██████╔╝
╚══════╝     ╚═════╝    ╚═╝  ╚═╝     ╚═════╝ 

``` 
