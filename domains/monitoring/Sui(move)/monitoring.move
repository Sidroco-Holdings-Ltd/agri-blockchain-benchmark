module my_first_package::monitoring {

    use std::address;
    use sui::transfer;
    use sui::event;
    use sui::tx_context::{Self, TxContext};
    use sui::object::UID;
    use sui::object;

    public struct Counter has key {
        id: UID,
        value: u64,
    }

    public struct IncrementEvent has copy, drop, store {
        sender: address,
        new_value: u64,
    }

    public entry fun initialize(ctx: &mut TxContext) {
        let id = object::new(ctx);
        let counter = Counter { id, value: 0 };
        transfer::share_object(counter);
    }

    public entry fun increment(counter: &mut Counter, ctx: &mut TxContext) {
        counter.value = counter.value + 1;
        let sender = tx_context::sender(ctx);
        let event = IncrementEvent {
            sender,
            new_value: counter.value,
        };
        event::emit(event);
    }

    public fun get_value(counter: &Counter): u64 {
        counter.value
    }
}