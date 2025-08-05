//! Lightweight message queue.

use std::collections::VecDeque;

/// A simple FIFO message queue.
pub struct MessageQueue<T> {
    queue: VecDeque<T>,
}

impl<T> MessageQueue<T> {
    /// Creates an empty queue.
    pub fn new() -> Self {
        Self {
            queue: VecDeque::new(),
        }
    }

    /// Adds a message to the back of the queue.
    pub fn enqueue(&mut self, msg: T) {
        self.queue.push_back(msg);
    }

    /// Removes and returns a message from the front of the queue.
    pub fn dequeue(&mut self) -> Option<T> {
        self.queue.pop_front()
    }

    /// Returns the number of messages in the queue.
    pub fn len(&self) -> usize {
        self.queue.len()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn queue_works() {
        let mut q = MessageQueue::new();
        q.enqueue(1);
        q.enqueue(2);
        assert_eq!(q.len(), 2);
        assert_eq!(q.dequeue(), Some(1));
        assert_eq!(q.dequeue(), Some(2));
        assert_eq!(q.dequeue(), None);
    }
}
