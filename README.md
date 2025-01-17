# ⚡️ SHM: System Hardware Monitor ⚡️

🚀 SHM (System Hardware Monitor) is a user-friendly, command-line tool designed to monitor system metrics on Linux machines. It utilizes a simple library: `sysmetricslib` library ([https://github.com/majdif47/sysmetricslib](https://github.com/majdif47/sysmetricslib)) to provide real-time insights into your system's performance. 

![shm](https://github.com/majdif47/shm/blob/main/250117_20h20m22s_screenshot.png)


**Features**

  - **CPU Monitoring:** View CPU usage 📈, core count 💪, thread count 👥, and frequency ⚡.
  - **Memory Monitoring:** Track total memory, used memory, available memory, and swap usage.
  - **Disk Monitoring:** Monitor disk usage (total, used, and available space) 💾 
  - **Network Monitoring:** 🌐information about network interfaces, including their state, speed 🚀, and data transfer statistics 📈📉.
  - **General Info Tab:** (Planned) Access essential system information in a dedicated tab.
  - **GPU Monitoring:** (Planned) 🎮 Keep an eye on your GPU's performance metrics 📊.
  - **Simple Task Manager:** (Planned) 🛠️ Manage running tasks directly from the SHM interface.

**Getting Started**

1.  **Prerequisites:**

      - Ensure you have Go installed on your system ([https://golang.org/doc/install](https://www.google.com/url?sa=E&source=gmail&q=https://golang.org/doc/install)).
      - Clone this repository: `git clone https://github.com/your-username/shm.git`
      - Navigate to the project directory: `cd shm`

2.  **Installation:**

      - Build the SHM executable: `go build -o shm .`

3.  **Running SHM:**

      - Execute the built binary: `./shm`

**Usage**

SHM presents a tabbed interface for navigating between different monitoring sections. Use the `Tab` key to switch between tabs and the `Up` and `Down` arrow keys to move within a table.

**Future Enhancements**
  - General information tab with system details.
  - GPU monitoring capabilities 🎮.
  - Simple task management features 🛠️.

**Contributing**
 ✨ Feel free to fork the repository, create pull requests with your improvements, and raise issues for any bugs or suggestions.
 
**Disclaimer**
This tool is currently in development and may not be fully optimized. Use it at your own discretion.
  - SHM is currently implemented for Linux systems only. will add Windows and MacOS.
  - The `sysmetricslib` library is a dependency of this project.
