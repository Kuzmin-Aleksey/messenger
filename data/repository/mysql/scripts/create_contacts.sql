create table if not exists contacts
(
    id              int auto_increment
        primary key,
    user_id         int         not null,
    contact_user_id int         not null,
    contact_name    varchar(45) not null,
    constraint users_id_idx
        unique (user_id, contact_user_id),
    constraint contacts_user_key
        foreign key (user_id) references users (id)
            on delete cascade
);