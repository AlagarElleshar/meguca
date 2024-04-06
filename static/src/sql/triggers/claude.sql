create or replace function after_claude_update()
returns trigger as $$
declare
    v_op bigint;
begin
    SELECT p.op into v_op
    FROM claude
             inner join public.posts p on claude.id = p.claude_id
    where claude.id = new.id;

    perform bump_thread(v_op, false);
    return null;
end;
$$ language plpgsql;