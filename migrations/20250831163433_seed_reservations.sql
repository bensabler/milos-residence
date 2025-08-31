-- +goose Up
-- +goose StatementBegin

-- Insert sample reservations with realistic data spread across past, present, and future
-- Room IDs: 1=Golden Haybeam Loft, 2=Window Perch Theater, 3=Laundry-Basket Nook
INSERT INTO reservations (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at, processed) VALUES
-- Past reservations (processed)
('Sarah', 'Johnson', 'sarah.johnson@email.com', '555-0101', '2024-12-01', '2024-12-03', 1, '2024-11-25 10:00:00', '2024-12-04 15:30:00', 1),
('Mike', 'Chen', 'mike.chen@email.com', '555-0102', '2024-12-05', '2024-12-08', 2, '2024-11-28 14:15:00', '2024-12-09 09:20:00', 1),
('Emma', 'Rodriguez', 'emma.rodriguez@email.com', '555-0103', '2024-12-10', '2024-12-12', 3, '2024-12-01 16:45:00', '2024-12-13 11:00:00', 1),
('James', 'Thompson', 'james.thompson@email.com', '555-0104', '2024-12-15', '2024-12-18', 1, '2024-12-05 09:30:00', '2024-12-19 14:15:00', 1),
('Lisa', 'Anderson', 'lisa.anderson@email.com', '555-0105', '2024-12-20', '2024-12-23', 2, '2024-12-10 13:20:00', '2024-12-24 10:45:00', 1),
('David', 'Wilson', 'david.wilson@email.com', '555-0106', '2024-12-27', '2024-12-30', 3, '2024-12-15 11:10:00', '2024-12-31 16:30:00', 1),
('Amy', 'Garcia', 'amy.garcia@email.com', '555-0107', '2025-01-02', '2025-01-05', 1, '2024-12-20 15:40:00', '2025-01-06 12:20:00', 1),
('Robert', 'Martinez', 'robert.martinez@email.com', '555-0108', '2025-01-08', '2025-01-11', 2, '2024-12-25 08:25:00', '2025-01-12 09:15:00', 1),
('Jennifer', 'Taylor', 'jennifer.taylor@email.com', '555-0109', '2025-01-15', '2025-01-17', 3, '2025-01-01 12:50:00', '2025-01-18 14:40:00', 1),
('Chris', 'Brown', 'chris.brown@email.com', '555-0110', '2025-01-20', '2025-01-23', 1, '2025-01-05 10:35:00', '2025-01-24 11:25:00', 1),

-- Recent/current reservations (mix of processed and unprocessed)
('Maria', 'Davis', 'maria.davis@email.com', '555-0111', '2025-01-25', '2025-01-28', 2, '2025-01-10 14:20:00', '2025-01-10 14:20:00', 1),
('Kevin', 'Miller', 'kevin.miller@email.com', '555-0112', '2025-02-01', '2025-02-04', 3, '2025-01-15 16:15:00', '2025-01-15 16:15:00', 0),
('Sophie', 'White', 'sophie.white@email.com', '555-0113', '2025-02-06', '2025-02-09', 1, '2025-01-20 09:45:00', '2025-01-20 09:45:00', 0),
('Daniel', 'Jones', 'daniel.jones@email.com', '555-0114', '2025-02-12', '2025-02-15', 2, '2025-01-25 13:30:00', '2025-01-25 13:30:00', 0),
('Rachel', 'Moore', 'rachel.moore@email.com', '555-0115', '2025-02-18', '2025-02-21', 3, '2025-01-30 11:55:00', '2025-01-30 11:55:00', 0),

-- Future reservations (unprocessed)
('Alex', 'Jackson', 'alex.jackson@email.com', '555-0116', '2025-02-25', '2025-02-28', 1, '2025-02-01 10:20:00', '2025-02-01 10:20:00', 0),
('Nicole', 'Lee', 'nicole.lee@email.com', '555-0117', '2025-03-03', '2025-03-06', 2, '2025-02-05 15:10:00', '2025-02-05 15:10:00', 0),
('Brandon', 'Clark', 'brandon.clark@email.com', '555-0118', '2025-03-10', '2025-03-13', 3, '2025-02-10 12:40:00', '2025-02-10 12:40:00', 0),
('Jessica', 'Lewis', 'jessica.lewis@email.com', '555-0119', '2025-03-15', '2025-03-18', 1, '2025-02-15 14:25:00', '2025-02-15 14:25:00', 0),
('Ryan', 'Walker', 'ryan.walker@email.com', '555-0120', '2025-03-22', '2025-03-25', 2, '2025-02-20 16:35:00', '2025-02-20 16:35:00', 0),
('Ashley', 'Hall', 'ashley.hall@email.com', '555-0121', '2025-03-28', '2025-03-31', 3, '2025-02-25 09:50:00', '2025-02-25 09:50:00', 0),
('Tyler', 'Young', 'tyler.young@email.com', '555-0122', '2025-04-02', '2025-04-05', 1, '2025-03-01 11:15:00', '2025-03-01 11:15:00', 0),
('Megan', 'King', 'megan.king@email.com', '555-0123', '2025-04-08', '2025-04-11', 2, '2025-03-05 13:45:00', '2025-03-05 13:45:00', 0),
('Jordan', 'Wright', 'jordan.wright@email.com', '555-0124', '2025-04-15', '2025-04-18', 3, '2025-03-10 15:20:00', '2025-03-10 15:20:00', 0),
('Hannah', 'Lopez', 'hannah.lopez@email.com', '555-0125', '2025-04-22', '2025-04-25', 1, '2025-03-15 10:30:00', '2025-03-15 10:30:00', 0),
('Nathan', 'Hill', 'nathan.hill@email.com', '555-0126', '2025-04-28', '2025-05-01', 2, '2025-03-20 12:10:00', '2025-03-20 12:10:00', 0),
('Samantha', 'Green', 'samantha.green@email.com', '555-0127', '2025-05-05', '2025-05-08', 3, '2025-03-25 14:55:00', '2025-03-25 14:55:00', 0),
('Justin', 'Adams', 'justin.adams@email.com', '555-0128', '2025-05-12', '2025-05-15', 1, '2025-03-30 16:40:00', '2025-03-30 16:40:00', 0),
('Olivia', 'Nelson', 'olivia.nelson@email.com', '555-0129', '2025-05-18', '2025-05-21', 2, '2025-04-05 09:25:00', '2025-04-05 09:25:00', 0),
('Ethan', 'Carter', 'ethan.carter@email.com', '555-0130', '2025-05-25', '2025-05-28', 3, '2025-04-10 11:05:00', '2025-04-10 11:05:00', 0);

-- Now create corresponding room restrictions for each reservation
-- We'll use a cursor-like approach with a temporary sequence to match reservation IDs
INSERT INTO room_restrictions (start_date, end_date, room_id, reservation_id, restriction_id, created_at, updated_at)
SELECT 
    r.start_date,
    r.end_date,
    r.room_id,
    r.id,
    1, -- restriction_id = 1 (Reservation type)
    r.created_at,
    r.updated_at
FROM reservations r
WHERE r.email IN (
    'sarah.johnson@email.com', 'mike.chen@email.com', 'emma.rodriguez@email.com', 'james.thompson@email.com',
    'lisa.anderson@email.com', 'david.wilson@email.com', 'amy.garcia@email.com', 'robert.martinez@email.com',
    'jennifer.taylor@email.com', 'chris.brown@email.com', 'maria.davis@email.com', 'kevin.miller@email.com',
    'sophie.white@email.com', 'daniel.jones@email.com', 'rachel.moore@email.com', 'alex.jackson@email.com',
    'nicole.lee@email.com', 'brandon.clark@email.com', 'jessica.lewis@email.com', 'ryan.walker@email.com',
    'ashley.hall@email.com', 'tyler.young@email.com', 'megan.king@email.com', 'jordan.wright@email.com',
    'hannah.lopez@email.com', 'nathan.hill@email.com', 'samantha.green@email.com', 'justin.adams@email.com',
    'olivia.nelson@email.com', 'ethan.carter@email.com'
);

-- Add a few owner blocks for maintenance periods
INSERT INTO room_restrictions (start_date, end_date, room_id, restriction_id, created_at, updated_at) VALUES
('2025-06-01', '2025-06-03', 1, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP), -- Owner block for Golden Haybeam Loft
('2025-06-15', '2025-06-17', 2, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP), -- Owner block for Window Perch Theater  
('2025-07-01', '2025-07-03', 3, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP); -- Owner block for Laundry-Basket Nook

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the seeded room restrictions first (due to foreign key constraints)
DELETE FROM room_restrictions 
WHERE reservation_id IN (
    SELECT id FROM reservations 
    WHERE email IN (
        'sarah.johnson@email.com', 'mike.chen@email.com', 'emma.rodriguez@email.com', 'james.thompson@email.com',
        'lisa.anderson@email.com', 'david.wilson@email.com', 'amy.garcia@email.com', 'robert.martinez@email.com',
        'jennifer.taylor@email.com', 'chris.brown@email.com', 'maria.davis@email.com', 'kevin.miller@email.com',
        'sophie.white@email.com', 'daniel.jones@email.com', 'rachel.moore@email.com', 'alex.jackson@email.com',
        'nicole.lee@email.com', 'brandon.clark@email.com', 'jessica.lewis@email.com', 'ryan.walker@email.com',
        'ashley.hall@email.com', 'tyler.young@email.com', 'megan.king@email.com', 'jordan.wright@email.com',
        'hannah.lopez@email.com', 'nathan.hill@email.com', 'samantha.green@email.com', 'justin.adams@email.com',
        'olivia.nelson@email.com', 'ethan.carter@email.com'
    )
);

-- Remove owner block restrictions
DELETE FROM room_restrictions 
WHERE restriction_id = 2 
AND start_date IN ('2025-06-01', '2025-06-15', '2025-07-01');

-- Remove the seeded reservations
DELETE FROM reservations 
WHERE email IN (
    'sarah.johnson@email.com', 'mike.chen@email.com', 'emma.rodriguez@email.com', 'james.thompson@email.com',
    'lisa.anderson@email.com', 'david.wilson@email.com', 'amy.garcia@email.com', 'robert.martinez@email.com',
    'jennifer.taylor@email.com', 'chris.brown@email.com', 'maria.davis@email.com', 'kevin.miller@email.com',
    'sophie.white@email.com', 'daniel.jones@email.com', 'rachel.moore@email.com', 'alex.jackson@email.com',
    'nicole.lee@email.com', 'brandon.clark@email.com', 'jessica.lewis@email.com', 'ryan.walker@email.com',
    'ashley.hall@email.com', 'tyler.young@email.com', 'megan.king@email.com', 'jordan.wright@email.com',
    'hannah.lopez@email.com', 'nathan.hill@email.com', 'samantha.green@email.com', 'justin.adams@email.com',
    'olivia.nelson@email.com', 'ethan.carter@email.com'
);

-- +goose StatementEnd