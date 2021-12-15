-- Functions
CREATE OR REPLACE FUNCTION updateDailyUserBalance(yesterdayStart DATE, yesterdayEnd TIMESTAMP with time zone, now TIMESTAMP with time zone) RETURNS SETOF record AS
$$
DECLARE
    r record;
    _id uuid;
BEGIN
    FOR r IN
        SELECT
            u.payment_cycle_id,
            u.id as user_id,
            -dp.amount as balance,
            'DAY' as balance_type,
            dp.currency,
            yesterdayStart as day,
            now as created_at$
        FROM daily_payment dp
                 INNER JOIN users u ON u.payment_cycle_id = dp.payment_cycle_id
                 INNER JOIN sponsor_event s ON s.user_id = u.id
                 INNER JOIN payment_cycle pc ON u.payment_cycle_id = pc.id
        WHERE pc.days_left > 0
          AND (EXTRACT(epoch from age(LEAST(yesterdayEnd, s.unsponsor_at), GREATEST(yesterdayStart, s.sponsor_at)))/3600)::bigInt >= 24
        ORDER BY u.id, dp.days_left asc
        LOOP
            if _id = r.payment_cycle_id then
                continue;
            end if;
            _id = r.payment_cycle_id;
            RETURN NEXT r;
        END LOOP;
    RETURN;
END;
$$
LANGUAGE plpgsql;
