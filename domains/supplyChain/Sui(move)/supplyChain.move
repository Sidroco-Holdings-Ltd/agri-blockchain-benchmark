module my_first_package::supply_chain {
    use sui::object;
    use sui::transfer;
    use sui::clock;
    use sui::clock::Clock;
    use sui::tx_context;
    use std::vector;
    use std::string;
    use std::option;
    use sui::tx_context::TxContext;

    public enum Status has copy, store, drop {
        Manufactured,
        Shipped,
        Received,
        Delivered,
    }

    public struct Product has key, store {
        id: object::UID,
        name: string::String,
        origin: string::String,
        destination: string::String,
        status: Status,
        owner: address,
        manufacturer: address,
        shipper: option::Option<address>,
        retailer: option::Option<address>,
        timestamp: u64,
    }

    public struct SupplyChainStorage has key {
        id: object::UID,
        products: vector<Product>,
        product_count: u64,
        manufacturers: vector<address>,
        shippers: vector<address>,
        retailers: vector<address>,
    }

    public entry fun initialize(ctx: &mut TxContext) {
        let storage = SupplyChainStorage {
            id: object::new(ctx),
            products: vector::empty(),
            product_count: 0,
            manufacturers: vector::empty(),
            shippers: vector::empty(),
            retailers: vector::empty(),
        };
        transfer::share_object(storage); 
    }

    public entry fun assign_roles(
        storage: &mut SupplyChainStorage,
        manufacturer: address,
        shipper: address,
        retailer: address,
    ) {
        assert!(storage.product_count >= 0, 100); 

        vector::push_back(&mut storage.manufacturers, manufacturer);
        vector::push_back(&mut storage.shippers, shipper);
        vector::push_back(&mut storage.retailers, retailer);
    }

    public entry fun add_product(
        storage: &mut SupplyChainStorage,
        name: string::String,
        origin: string::String,
        destination: string::String,
        clock: &Clock,
        ctx: &mut TxContext,
    ) {
        let sender = tx_context::sender(ctx);

        assert!(vector::contains(&storage.manufacturers, &sender), 0);

        let product = Product {
            id: object::new(ctx),
            name,
            origin,
            destination,
            status: Status::Manufactured,
            owner: sender,
            manufacturer: sender,
            shipper: option::none(),
            retailer: option::none(),
            timestamp: clock::timestamp_ms(clock),
        };

        vector::push_back(&mut storage.products, product);
        storage.product_count = storage.product_count + 1;
    }

    public entry fun update_status(
        storage: &mut SupplyChainStorage,
        product_index: u64,
        new_status_code: u8,
        clock: &Clock,
        ctx: &mut TxContext,
    ) {
        let new_status = match (new_status_code) {
            0u8 => Status::Manufactured,
            1u8 => Status::Shipped,
            2u8 => Status::Received,
            3u8 => Status::Delivered,
            _ => abort(6), // Invalid status code
        };

        let sender = tx_context::sender(ctx);

        assert!(product_index < vector::length(&storage.products), 1);
        let product = vector::borrow_mut(&mut storage.products, product_index);

        match (new_status) {
            Status::Shipped => {
                assert!(vector::contains(&storage.shippers, &sender), 2);
                product.shipper = option::some(sender);
            },
            Status::Received | Status::Delivered => {
                assert!(vector::contains(&storage.retailers, &sender), 3);
                if (new_status == Status::Delivered) {
                    assert!(product.status == Status::Received, 4); // Ensure valid transition
                };
                product.retailer = option::some(sender);
            },
            Status::Manufactured => { /* No special action */ }
        };

        product.status = new_status;
        product.timestamp = clock::timestamp_ms(clock);
    }

    public fun get_product(storage: &SupplyChainStorage, product_index: u64): &Product {
        assert!(product_index < vector::length(&storage.products), 5);
        vector::borrow(&storage.products, product_index)
    }
}
