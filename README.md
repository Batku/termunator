# Termunator (name WiP)

Termunator is a modern SSH and SFTP client built with Go, Wails, and Svelte. It provides a streamlined interface for managing remote connections, file transfers, and host configurations.

## Warning

This project is incomplete

## Showcase
<img width="1080" height="727" alt="image" src="https://github.com/user-attachments/assets/d0b432a0-a0af-4339-a628-462fec415020" />
<img width="1080" height="600" alt="image" src="https://github.com/user-attachments/assets/d8a35cec-f716-46bb-9288-6c39fe0b4f11" />
<img width="1080" height="894" alt="image" src="https://github.com/user-attachments/assets/574263af-065d-45de-97cb-c2869e70b7fc" />




## Features

- **SSH Client:** Connect securely to remote servers using SSH.
- **SFTP Support:** Transfer files between your local machine and remote hosts with ease.
- **Host Management:** Save and organize your frequently used hosts and connection settings.
- **Flexible Storage Options:** Choose how your hosts and settings are stored:
  - **Local Storage:** Keep your data on your device for privacy and offline access.
  - **Self-Hosted Backend:** Connect to your own backend server for centralized management.
  - **Subscription-Based Cloud:** (Coming soon) Use our managed cloud service for seamless sync and access across devices.

## Getting Started

Termunator is currently in development. There are no build guides.

## Contributing

Contributions are welcome! Please see the repository for guidelines and ways to get involved (This line is here for when its more complete...)

## License

This project is open source and available under the terms of the license found in this repository.

## Attributions & Licenses

Termunator uses the following open source libraries and assets:

### Frontend (JavaScript/TypeScript)

- [Svelte](https://svelte.dev/) — UI framework
- [Vite](https://vitejs.dev/) — Build tool
- [Tailwind CSS](https://tailwindcss.com/) — Utility-first CSS framework
- [lucide-svelte](https://lucide.dev/) — Icon library (MIT License)
- [xterm.js](https://xtermjs.org/) and addons — Terminal emulator
- [PostCSS](https://postcss.org/) — CSS processing
- [Autoprefixer](https://github.com/postcss/autoprefixer) — CSS vendor prefixing

### Backend (Go)

- [Wails](https://wails.io/) — Go desktop app framework
- [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) — Cryptography
- [github.com/pkg/sftp](https://github.com/pkg/sftp) — SFTP protocol
- [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) — SQLite driver
- [github.com/google/uuid](https://github.com/google/uuid) — UUID generation

Other dependencies are listed in `package.json` and `go.mod`. Please refer to those files for the full list and their respective licenses.

If you use Termunator in a way that redistributes it, ensure you comply with the licenses of these dependencies. For icon and font attributions, see the respective project documentation.



