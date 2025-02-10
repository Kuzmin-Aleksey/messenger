create table if not exists user_2_chat
(
    id      int auto_increment
        primary key,
    user_id int not null,
    chat_id int not null,
    constraint chat_key
        foreign key (chat_id) references chats (id)
            on delete cascade,
    constraint user_key
        foreign key (user_id) references users (id)
            on delete cascade
);