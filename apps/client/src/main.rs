use tokio::net::TcpStream;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::time::{sleep, Duration};
use std::error::Error;

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let mut stream = TcpStream::connect("127.0.0.1:8080").await?;
    println!("Connected to server");

    let mut buffer = [0; 1024];
    let mut count = 0;

    loop {
        count += 1;
        let ping = format!("ping {}", count);
        stream.write_all(ping.as_bytes()).await?;
        println!("Sent: {}", ping);

        let n = stream.read(&mut buffer).await?;
        if n == 0 {
            println!("Server disconnected");
            break;
        }
        
        let response = String::from_utf8_lossy(&buffer[..n]);
        println!("Received: {}", response);

        sleep(Duration::from_secs(1)).await;
    }

    Ok(())
} 