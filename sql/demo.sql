INSERT INTO rounds (round_id, name, description, time, expires, password) VALUES
('6d2fb784-1732-4350-872c-26d47793109e', 'test', 'Boire un coup pour tester', '2026-09-12T21:00', '2026-12-13T03:00', 'ql+gU6CkIasPSrblo+78FO7J/WnR4fZ+ycK8weyg/BHLhB9N8XPa23oD5nxJmA79/c2sIgPZF5g='); -- binouze

INSERT INTO orders (order_id, round_id, name, password) VALUES
('03ee6484-9678-4e1b-8080-0995634252c2', '6d2fb784-1732-4350-872c-26d47793109e', 'Jérôme', '8U1D+Q/Xx97GDMGgy/l97s4N9LhE1djdME8mnTwqFi82d8dFxdav8gbXJgAuSzt0vkTkn4/7bVE='), -- testtest
('b8eba9ab-3e4b-4e86-a1f7-f1279551acb3', '6d2fb784-1732-4350-872c-26d47793109e', 'Michel', 'K5EoaeD+Y6GkwGLxyrWXyxS/BqYca4EYiWZMrvh+YjQ8NkIhxhmHcsGD6AsnRyqPmYlNf+miXNs='); -- motdepasse

INSERT INTO drinks (drink_id, order_id, name, quantity) VALUES
('1e2dbba3-7f6b-4acb-854e-a6c66f4ac174', '03ee6484-9678-4e1b-8080-0995634252c2', 'Bière d''été', 2),
('e9950b2a-9b8e-4005-9c1a-5c28dd562f02', '03ee6484-9678-4e1b-8080-0995634252c2', 'Cacahuètes', 1),
('4171b255-9c80-44ec-83ad-a330a58f9fc3', '03ee6484-9678-4e1b-8080-0995634252c2', 'Coca', 1),
('02f7c720-9400-4b4d-8544-103ff0702f88', 'b8eba9ab-3e4b-4e86-a1f7-f1279551acb3', 'Coca', 2),
('34335de5-d064-47ce-b63b-075f894ecab3', 'b8eba9ab-3e4b-4e86-a1f7-f1279551acb3', 'Chips', 1);