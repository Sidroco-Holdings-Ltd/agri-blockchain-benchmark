module my_first_package::defi_protocol {
    use sui::object::UID;
    use sui::tx_context::TxContext;
    use sui::transfer;
    use sui::event;
    use std::address;

    public struct StakingPool has key {
        id: UID,
        owner: address,
        total_staked: u64,
        reward_rate: u64,
    }

    public struct Stake has key {
        id: UID,
        owner: address,
        staked_amount: u64,
        reward_accumulated: u64,
    }

    public struct StakeEvent has copy, drop, store {
        sender: address,
        amount: u64,
    }

    public struct WithdrawEvent has copy, drop, store {
        sender: address,
        withdrawn_amount: u64,
        reward_amount: u64,
    }

    public entry fun initialize_pool(
        reward_rate: u64,
        ctx: &mut TxContext
    ) {
        let id = object::new(ctx);
        let owner = tx_context::sender(ctx);
        let pool = StakingPool {
            id,
            owner,
            total_staked: 0,
            reward_rate,
        };
        transfer::share_object(pool);
    }

    public entry fun stake_tokens(
        pool: &mut StakingPool,
        stake_amount: u64,
        ctx: &mut TxContext
    ) {
        assert!(stake_amount > 0, 0); // Error code 0: Invalid amount

        let id = object::new(ctx);
        let owner = tx_context::sender(ctx);
        let stake = Stake {
            id,
            owner,
            staked_amount: stake_amount,
            reward_accumulated: 0,
        };

        pool.total_staked = pool.total_staked + stake_amount;

        event::emit(StakeEvent {
            sender: owner,
            amount: stake_amount,
        });

        transfer::transfer(stake, owner);
    }

    public entry fun withdraw(
        pool: &mut StakingPool,
        stake: &mut Stake,
        ctx: &mut TxContext
    ) {
        let sender = tx_context::sender(ctx);
        assert!(sender == stake.owner, 1); // Error code 1: Unauthorized

        let reward = calculate_reward(
            stake.staked_amount,
            pool.reward_rate,
        );

        let total_withdrawn = stake.staked_amount + reward;

        pool.total_staked = pool.total_staked - stake.staked_amount;

        stake.staked_amount = 0;
        stake.reward_accumulated = 0;

        event::emit(WithdrawEvent {
            sender,
            withdrawn_amount: total_withdrawn,
            reward_amount: reward,
        });
    }

    public fun calculate_reward(
        staked_amount: u64,
        reward_rate: u64
    ): u64 {
        staked_amount * reward_rate
    }
}
