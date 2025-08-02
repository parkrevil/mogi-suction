use shared;
use tokio::net::{TcpListener, TcpStream};
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use std::error::Error;

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let listener = TcpListener::bind("127.0.0.1:8080").await?;
    println!("TCP Server listening on 127.0.0.1:8080");

    loop {
        let (socket, addr) = listener.accept().await?;
        println!("New connection from: {}", addr);
        
        tokio::spawn(async move {
            if let Err(e) = handle_connection(socket).await {
                eprintln!("Error handling connection: {}", e);
            }
        });
    }
}

async fn handle_connection(mut socket: TcpStream) -> Result<(), Box<dyn Error>> {
    let mut buffer = [0; 1024];
    
    loop {
        let n = socket.read(&mut buffer).await?;
        if n == 0 {
            println!("Client disconnected");
            break;
        }
        
        let received = String::from_utf8_lossy(&buffer[..n]);
        println!("Received: {}", received.trim());
        
        // Echo back the received message
        let response = format!("Server received: {}", received.trim());
        socket.write_all(response.as_bytes()).await?;
    }
    
    Ok(())
} 