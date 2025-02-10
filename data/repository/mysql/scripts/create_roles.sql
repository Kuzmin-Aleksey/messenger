create table if not exists roles
(
    id       int auto_increment
        primary key,
    user_id  int                          not null,
    group_id int                          not null,
    role     varchar(45) default 'member' not null,
    constraint roles_group_key
        foreign key (group_id) references `groups` (id)
            on delete cascade,
    constraint roles_user_key
        foreign key (user_id) references users (id)
            on delete cascade
);