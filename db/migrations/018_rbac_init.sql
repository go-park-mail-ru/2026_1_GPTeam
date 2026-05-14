create role app_service_role;
create role file_service_role;
create role ai_service_role;
create role auth_service_role;

grant usage on schema public to app_service_role, auth_service_role;
grant insert, delete, select on jwt to auth_service_role;
grant select on pg_enum to app_service_role;
grant insert, select, update on "user", budget, transaction, account, support to app_service_role;
grant insert, select, delete on budget_category, account_user to app_service_role;

---- create above / drop below ----

revoke usage on schema public from app_service_role, auth_service_role;
revoke insert, delete, select on jwt from auth_service_role;
revoke select on pg_enum from app_service_role;
revoke insert, select, update on "user", budget, transaction, account, support from app_service_role;
revoke insert, select, delete on budget_category, account_user from app_service_role;
