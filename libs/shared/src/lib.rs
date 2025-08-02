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