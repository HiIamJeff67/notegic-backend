-- 0000_billing_plan_seed.up.sql
INSERT INTO "BillingPlanTable" (
    id, 
    product_id, 
    name, 
    status,
    interval_unit,
    price, 
    currency_code, 
    updated_at,
    created_at
) VALUES
('P-FREE-PLAN-ID',             'NO-PRODUCT',              'Notegic Monthly Free Plan',       'ACTIVE', 'MONTH', 0,      'USD', NOW(), NOW()),
('P-4LN51972TD528344JNHJX5NQ', 'PROD-22007905DJ014973Y',  'Notegic Monthly Pro Plan',        'ACTIVE', 'MONTH', 4.99,   'USD', NOW(), NOW()),
('P-9MB559415V1980509NHJX7LY', 'PROD-22007905DJ014973Y',  'Notegic Yearly Pro Plan',         'ACTIVE', 'YEAR',  49.99,  'USD', NOW(), NOW()),
('P-351611974A6912332NHJYECY', 'PROD-22007905DJ014973Y',  'Notegic Monthly Premium Plan',    'ACTIVE', 'MONTH', 9.99,   'USD', NOW(), NOW()),
('P-84627481GN3337838NHJYEJA', 'PROD-22007905DJ014973Y',  'Notegic Yearly Premium Plan',     'ACTIVE', 'YEAR',  99.99,  'USD', NOW(), NOW()),
('P-3B912255TH6394814NHJYEMA', 'PROD-22007905DJ014973Y',  'Notegic Monthly Ultimate Plan',   'ACTIVE', 'MONTH', 19.99,  'USD', NOW(), NOW()),
('P-4WS50500MM359840MNHJYEOY', 'PROD-22007905DJ014973Y',  'Notegic Yearly Ultimate Plan',    'ACTIVE', 'YEAR',  199.99, 'USD', NOW(), NOW()),
('P-9XP882067J683411BNHJYERI', 'PROD-22007905DJ014973Y',  'Notegic Monthly Enterprise Plan', 'ACTIVE', 'MONTH', 49.99,  'USD', NOW(), NOW()),
('P-2PT73314S5217944VNHJYEYI', 'PROD-22007905DJ014973Y',  'Notegic Yearly Enterprise Plan',  'ACTIVE', 'YEAR',  499.99, 'USD', NOW(), NOW())
ON CONFLICT (name) DO UPDATE SET
    id = EXCLUDED.id, 
    product_id = EXCLUDED.product_id, 
    name = EXCLUDED.name,
    status = EXCLUDED.status,
    interval_unit = EXCLUDED.interval_unit, 
    price = EXCLUDED.price, 
    currency_code = EXCLUDED.currency_code,
    updated_at = NOW();