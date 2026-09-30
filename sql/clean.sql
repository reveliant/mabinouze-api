CREATE TEMPORARY TABLE expired_rounds
    AS SELECT round_id, name, time, locked
    FROM rounds
    WHERE rounds.expires < CURRENT_TIMESTAMP;

SELECT round_id, name FROM expired_rounds;

INSERT INTO stats (time, tipplers, drinks)
    SELECT expired_rounds.time, COUNT(orders.order_id) as tipplers, COALESCE(SUM(drinks.quantity), 0)
    FROM expired_rounds
    LEFT JOIN orders USING (round_id)
    LEFT JOIN drinks USING (order_id)
    GROUP BY expired_rounds.time;

DELETE FROM drinks WHERE order_id IN (
    SELECT order_id
    FROM orders, expired_rounds
    WHERE orders.round_id = expired_rounds.round_id
);

DELETE FROM orders WHERE round_id IN (
    SELECT round_id
    FROM expired_rounds
);

DELETE FROM rounds WHERE round_id IN (
    SELECT round_id
    FROM expired_rounds
    WHERE locked IS NOT TRUE
);

DROP TABLE expired_rounds;