create table if not exists messages
(
    id      int auto_increment
        primary key,
    chat_id int                not null,
    user_id int                not null,
    value   text charset utf32 not null,
    time    datetime           not null,
    constraint messages_chat_key
        foreign key (chat_id) references chats (id)
            on delete cascade,
    constraint messages_user_key
        foreign key (user_id) references users (id)
            on delete cascade
);