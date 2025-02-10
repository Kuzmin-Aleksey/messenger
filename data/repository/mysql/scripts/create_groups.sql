create table if not exists `groups`
(
    id      int auto_increment
        primary key,
    chat_id int         not null,
    name    varchar(45) not null,
    constraint chat_id_UNIQUE
        unique (chat_id),
    constraint groups_chat_key
        foreign key (chat_id) references chats (id)
            on delete cascade
);