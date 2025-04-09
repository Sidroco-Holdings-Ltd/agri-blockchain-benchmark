module my_first_package::data_storage {
    use sui::object::UID;
    use sui::tx_context::TxContext;
    use sui::transfer;
    use std::vector;
    use std::string;

    public struct DataItem has key {
        id: UID,
        owner: address,
        data: vector<u8>,
    }

    public entry fun initialize(
        data: vector<u8>,
        ctx: &mut TxContext
    ) {
        let id = object::new(ctx);
        let owner = tx_context::sender(ctx);
        let data_item = DataItem {
            id,
            owner,
            data,
        };
        transfer::share_object(data_item);
    }

    public entry fun update_data(
        data_item: &mut DataItem,
        new_data: vector<u8>,
        ctx: &TxContext
    ) {
        let sender = tx_context::sender(ctx);
        assert!(sender == data_item.owner, 0); // Error code 0: Unauthorized

        data_item.data = new_data;
    }

    public fun get_data(data_item: &DataItem): vector<u8> {
        data_item.data
    }
}
