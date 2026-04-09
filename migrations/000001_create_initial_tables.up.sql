create table if not exists users(
    id serial primary key, 
    balance int default 0,
    withdrawn int default 0, 
    login varchar(255) not null, 
    password varchar(255) not null);


create table if not exists orders(
    id serial primary key, 
    user_id int not null, 
    number varchar(10), 
    status varchar(15) default 'NEW' CHECK(status IN('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')), 
    created_at timestamptz DEFAULT NOW(),
    updated_at timestamptz DEFAULT NOW(),
    points int DEFAULT 0,
    FOREIGN KEY(user_id) REFERENCES users(id) );


create table if not exists withdrawals(
    id serial, 
    user_id int not null, 
    order_number text, 
    sum int, 
    processed_at TIMESTAMPTZ default now());