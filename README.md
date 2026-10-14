# High-Performance Packet Sniffer & Network Analyzer

A lightweight, low-level CLI tool and live terminal dashboard written in Go that intercepts, decodes, and reassembles network traffic in real time.

Instead of relying on API proxies, it hooks directly into your system's network driver layer (libpcap / BPF), reassembles TCP streams in memory, and reconstructs complete HTTP transactions - pairing every client request with its corresponding server response, headers and JSON bodies included - then presents them in an interactive TUI.

## Technologies

* **Packet Capture & Decoding:** [`github.com/google/gopacket`](https://github.com/google/gopacket) (`pcap`, `layers`, `reassembly`)

* **Terminal UI:**

  * [`charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea) (Elm-architecture TUI framework)

  * [`charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss) (Terminal styling & layout)

  * [`charmbracelet/bubbles`](https://github.com/charmbracelet/bubbles) (Table component)

* **Native OS Drivers:** `libpcap` (macOS/Linux) or `Npcap` (Windows)

## Prerequisites

Because packet capture interfaces directly with kernel network drivers, you must have the appropriate native libraries installed:

* **macOS:** Install `libpcap` via Homebrew:

  ```
  brew install libpcap
  
  ```

* **Linux:** Install `libpcap-dev`:

  ```
  sudo apt install libpcap-dev   # Ubuntu/Debian
  
  ```

* **Windows:** Install [Npcap](https://npcap.com/) (select *Install Npcap in WinPcap API-compatible Mode* during setup).

## Installation & Build

1. Clone the repository:

   ```
   git clone https://github.com/your-username/packet-sniffer.git
   cd packet-sniffer
   
   ```

2. Download dependencies:

   ```
   go mod download
   
   ```

3. Build the executable binary:

   ```
   go build -o sniffer ./cmd/main.go
   
   ```

## Usage

Packet capture requires elevated privileges to access network devices (`/dev/bpf*` on macOS or Npcap drivers on Windows).

### Running on macOS / Linux

Run the compiled binary with `sudo`:

```
# Capture traffic on primary Wi-Fi / Ethernet interface (default: en0)
sudo ./sniffer

# Specify a custom interface (e.g., loopback for localhost traffic)
sudo ./sniffer lo0

```

> **Note on Interface Selection:**
>
> * Use `en0` (or `eth0`) for external internet traffic.
>
> * Use `lo0` (or `lo`) to inspect local microservices running on `localhost` / `127.0.0.1`.

### Running on Windows

Open PowerShell or Command Prompt **as Administrator**:

```
.\sniffer.exe

```

## TUI Navigation

* **`↑` / `↓` / `k` / `j`**: Scroll through incoming HTTP requests in the live table.

* **`Enter`**: Inspect the selected request (displays headers, metadata, and formatted JSON body).

* **`Esc`**: Return to the live table view from the inspector panel.

* **`q` / `Ctrl+C`**: Quit the application.

## Testing the Sniffer

Because modern web browsers default to encrypted HTTPS (Port 443), run these commands in a separate terminal to generate plain HTTP traffic for testing:

### 1. Test Plaintext GET Request

```
curl -i http://neverssl.com

```

### 2. Test POST Request with JSON Body

```
curl -i -X POST http://httpbin.org/post \
  -H "Content-Type: application/json" \
  -d '{"user_id": 42, "status": "active", "payload": {"debug": true}}'

```

### 3. Test Local Traffic (Run sniffer on `lo0`)

```
# Terminal 1: Start local python server
python3 -m http.server 8080

# Terminal 2: Sniff loopback
sudo ./sniffer lo0

# Terminal 3: Send request
curl http://127.0.0.1:8080/

```