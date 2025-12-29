begin;

create table if not exists migrations(
  num int not null unique,
  title varchar(255) not null,
  applied bool
);

end;
