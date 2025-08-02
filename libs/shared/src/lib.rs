use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct Message {
    pub content: String,
}

pub fn create_message(content: &str) -> Message {
    Message {
        content: content.to_string(),
    }
}

// Ping-Pong 관련 구조체
#[derive(Debug, Serialize, Deserialize)]
pub struct Ping {
    pub id: u32,
    pub timestamp: u64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Pong {
    pub id: u32,
    pub timestamp: u64,
} 