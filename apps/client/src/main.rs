use pcap::{Capture, Device};
use serde::Serialize;
use std::error::Error;

#[derive(Debug, Serialize)]
struct PacketInfo {
    direction: String,
    hex_data: String,
}

fn capture_packets() -> Result<(), Box<dyn Error>> {
    let devices = Device::list()?;

    if devices.is_empty() {
        return Err("No available network interfaces found.".into());
    }

    let device = devices.first().unwrap();

    println!("Selected interface: {}", device.name);

    let mut cap = Capture::from_device(device.clone())?
        .promisc(true)
        .snaplen(65535)
        .open()?;

    cap.filter("port 16000", true)?;

    println!("Capturing packets...");

    let mut packet_count = 0;

    loop {
        match cap.next_packet() {
            Ok(packet) => {
                packet_count += 1;

                let packet_info = analyze_packet(&packet, packet_count);

                if let Ok(json) = serde_json::to_string_pretty(&packet_info) {
                    println!("패킷 #{}: {}", packet_count, json);
                }
            }
            Err(e) => {
                eprintln!("패킷 캡처 오류: {}", e);
                break;
            }
        }
    }

    Ok(())
}

fn analyze_packet(packet: &pcap::Packet, _packet_num: u32) -> PacketInfo {
    let hex_data = packet
        .data
        .iter()
        .map(|b| format!("{:02x}", b))
        .collect::<Vec<String>>()
        .join("");

    let mut direction = "unknown".to_string();

    if packet.data.len() >= 38 {
        let src_port = Some(u16::from_be_bytes([packet.data[34], packet.data[35]]));
        let dst_port = Some(u16::from_be_bytes([packet.data[36], packet.data[37]]));

        if src_port == Some(16000) {
            direction = "out".to_string();
        } else if dst_port == Some(16000) {
            direction = "in".to_string();
        }
    }

    PacketInfo {
        direction,
        hex_data: if hex_data.len() > 100 {
            format!("{}...", &hex_data[..100])
        } else {
            hex_data
        },
    }
}

fn main() -> Result<(), Box<dyn Error>> {
    match capture_packets() {
        Ok(_) => println!("Done"),
        Err(e) => eprintln!("Error: {}", e),
    }

    Ok(())
}
