create table if not exists chats
(
    id                int auto_increment
        primary key,
    type              varchar(32) default 'group'               not null,
    create_time       datetime    default CURRENT_TIMESTAMP     not null,
    last_message_time datetime    default '0001-01-01 00:00:00' not null
);