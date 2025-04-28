# Go-Raft

Go-Raft is a distributed consensus implementation based on the Raft protocol. It ensures consistency across a cluster of servers in a fault-tolerant manner.

## Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/longquan0104/go-raft
   cd go-raft
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the server:
   ```bash
   go run ./server -server-name=<server-name> -port=<port>
   ```
   - `<server-name>`: The name or ID of the server.
   - `<port>`: The port to send requests to the server.

   Example:
   ```bash
   go run ./server -server-name=serverOne -port=8001
   ```

4. Run the client:
   ```bash
   go run ./client <url-of-the-leader>
   ```
   - `<url-of-the-leader>`: The URL of the leader server.

   Example:
   ```bash
   go run ./client localhost:8001
   ```

   If no URL is provided:
   ```bash
   go run ./client
   ```
   The client will automatically use the leader URL from the predefined configuration file after leader election is completed.

## Usage

To run the servers, use the following aliases:
```bash
alias serverone="go run ./server -server-name=serverone -port=8001"
alias servertwo="go run ./server -server-name=servertwo -port=8002"
alias serverthr="go run ./server -server-name=serverthr -port=8003"
```

Start each server in a separate terminal window by running the respective alias commands.

## Testing

To test the functionality of the Raft implementation, interact with the servers using their respective ports (e.g., `8001`, `8002`, `8003`).

## Disclaimer

This project is created solely for learning purposes to better understand the Raft consensus algorithm. It is not intended for production use.

## Acknowledgments

This project is inspired by and adapted from [graft](https://github.com/varunu28/graft.git). Special thanks to the original author for their work on implementing the Raft protocol.

For better understanding:
- Read the [Raft paper](https://raft.github.io/raft.pdf).
- Watch this [course](https://youtu.be/rN6ma561tak?si=_FkLcJyXL--OlMUX).

Special thanks to my ex tech lead for guidance and support.
