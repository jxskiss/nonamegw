use std::marker::PhantomData;
use std::ptr::null_mut;
use std::rc::Rc;

// https://www.ralfj.de/projects/rust-101/part16.html

pub struct Node<T> {
    prev: NodePtr<T>,
    next: NodePtr<T>,
    value: Option<T>,
}

impl<T> Node<T> {
    pub fn null() -> Self {
        Self {
            prev: null_mut(),
            next: null_mut(),
            value: None,
        }
    }

    pub fn new(value: T) -> Self {
        Self {
            prev: null_mut(),
            next: null_mut(),
            value: Some(value),
        }
    }
}

pub type NodePtr<T> = *mut Node<T>;

pub fn deref_node(node: NodePtr<T>) -> T {

}

pub struct LinkedList<T> {
    root: NodePtr<T>,
    len: usize,
    _marker: PhantomData<T>,
}

impl<T> LinkedList<T> {
    pub fn new() -> Self {
        let root = Box::new(Node::null());
        let root_ptr = Box::into_raw(root);
        unsafe {
            (*root_ptr).prev = root_ptr;
            (*root_ptr).next = root_ptr;
        }
        Self {
            root: root_ptr,
            len: 0,
            _marker: PhantomData,
        }
    }

    pub fn is_empty(&self) -> bool {
        self.len == 0
    }

    pub fn len(&self) -> usize {
        self.len
    }

    pub fn front(&self) -> NodePtr<T> {
        if self.len == 0 {
            return null_mut()
        }
        self.root.next
    }

    pub fn back(&self) -> NodePtr<T> {
        if self.len == 0 {
            return null_mut()
        }
        self.root.prev
    }

    pub fn next(&self, node: NodePtr<T>) -> NodePtr<T> {
        if node.next == self.root {
            return null_mut()
        }
        node.next
    }

    #[inline(always)]
    unsafe fn _insert(&mut self, node: NodePtr<T>, at: NodePtr<T>) {
        let node = &mut *node;
        node.prev = at;
        node.next = (*at).next;
        (*node.prev).next = node;
        (*node.next).prev = node;
        self.len += 1;
    }

    pub fn push_front(&mut self, value: T) -> NodePtr<T> {
        let node = Box::new(Node::new(value));
        let node_ptr = Box::into_raw(node);
        unsafe {
            let root_next = (*self.root).next;
            self._insert(node_ptr, root_next);
        }
        node_ptr
    }

    pub fn push_back(&mut self, value: T) -> NodePtr<T> {
        let node = Box::new(Node::new(value));
        let node_ptr = Box::into_raw(node);
        unsafe {
            let root_prev = (*self.root).prev;
            self._insert(node_ptr, root_prev);
        }
        node_ptr
    }

    /// 删除传入的 NodePtr 对应的链表节点，节点被删除后，外部不应该以任何方式继续持有
    /// 或访问 NodePtr 指针。
    pub unsafe fn remove(&mut self, node: NodePtr<T>) -> Option<T> {
        let node_ref = &mut *node;
        (*node_ref.prev).next = node_ref.next;
        (*node_ref.next).prev = node_ref.prev;
        self.len -= 1;

        // free memory
        let node = Box::from_raw(node);
        node.value
    }

    #[inline(always)]
    unsafe fn _move(&mut self, node: NodePtr<T>, at: NodePtr<T>) {
        if node == at {
            return;
        }
        let node = &mut *node;
        (*node.prev).next = node.next;
        (*node.next).prev = node.prev;
        node.prev = at;
        node.next = (*at).next;
        (*node.prev).next = node;
        (*node.next).prev = node;
    }

    pub fn move_to_front(&mut self, node: NodePtr<T>) {
        unsafe {
            let root_next = (*self.root).next;
            self._move(node, root_next);
        }
    }

    pub fn move_to_back(&mut self, node: NodePtr<T>) {
        unsafe {
            let root_prev = (*self.root).prev;
            self._move(node, root_prev);
        }
    }
}

impl<T> Drop for LinkedList<T> {
    fn drop(&mut self) {
        unsafe {
            let mut next = (*self.root).next;
            for i in 0..self.len {
                let node = Box::from_raw(next);
                next = node.next;
                drop(node);
            }
            let root = Box::from_raw(self.root);
            drop(root);
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_doubly_linked_list() {
        {
            let mut list = LinkedList::new();
            let ptr1 = list.push_back(1);
            println!("len= {}, ptr1= {:?}", list.len, ptr1);
        }
        {
            let mut list = LinkedList::new();
            let ptr1 = list.push_front(2);
            let ptr2 = list.push_front(3);
            println!("len= {}, ptr1= {:?}, ptr2= {:?}", list.len, ptr1, ptr2);

            println!("{:?}", unsafe { list.remove(ptr1) });
            println!("{:?}", unsafe { list.remove(ptr2) });
            println!("len= {}, ptr1= {:?}, ptr2= {:?}", list.len, ptr1, ptr2);
        }

        // test drop node
        {
            let mut list = LinkedList::new();
            let obj = Rc::new(Node::new(1));
            assert_eq!(Rc::strong_count(&obj), 1);

            let ptr1 = list.push_front(Rc::clone(&obj));
            assert_eq!(Rc::strong_count(&obj), 2);

            unsafe { list.remove(ptr1) };
            assert_eq!(Rc::strong_count(&obj), 1)
        }
    }
}
