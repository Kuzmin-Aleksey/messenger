create table if not exists users
(
    id          int auto_increment
        primary key,
    phone       varchar(16)                        not null,
    password    tinyblob                           not null,
    name        varchar(45)                        not null,
    real_name   varchar(45)                        not null,
    show_phone  tinyint  default 1                 not null,
    last_online datetime default CURRENT_TIMESTAMP not null,
    confirmed   tinyint  default 0                 not null,
    constraint name_UNIQUE
        unique (name),
    constraint phone_UNIQUE
        unique (phone)
);
