use shared;
use tokio::net::TcpStream;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use std::error::Error;

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let mut stream = TcpStream::connect("127.0.0.1:8080").await?;
    println!("Connected to server at 127.0.0.1:8080");

    // Send a test message
    let message = "Hello from client!";
    stream.write_all(message.as_bytes()).await?;
    println!("Sent: {}", message);

    // Read response
    let mut buffer = [0; 1024];
    let n = stream.read(&mut buffer).await?;
    let response = String::from_utf8_lossy(&buffer[..n]);
    println!("Received: {}", response);

    Ok(())
} 