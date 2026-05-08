create role app_service_role;
create role file_service_role;
create role ai_service_role;
create role auth_service_role;

grant insert, delete, select on jwt to auth_service_role;
grant select on pg_enum to app_service_role;
grant insert, select, update on "user", budget, transaction, account, support to app_service_role;
grant insert, select, delete on budget_category, account_user to app_service_role;

---- create above / drop below ----

drop role if exists app_service_role;
drop role if exists file_service_role;
drop role if exists ai_service_role;
drop role if exists auth_service_role;
