use anyhow::Result;
use itoa::Buffer;
use rand::Rng;

fn main() -> Result<()> {
    let n: u8 = rand::thread_rng().r#gen();
    println!("Random number: {n}");

    let mut buf = Buffer::new();
    println!("Random number as string: {}", buf.format(n));

    Ok(())
}

#[cfg(test)]
mod tests {
    use futures::executor::block_on;
    use futures::future::ready;

    #[test]
    fn test_future() {
        let fut = ready(42);
        let val = block_on(fut);
        assert_eq!(val, 42);
    }
}
