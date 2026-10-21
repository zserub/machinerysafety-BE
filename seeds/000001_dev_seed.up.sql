-- =============================================================================
-- Dev Seed: Mock data for development and testing
-- NOT a production migration: only applied when SEEDS_PATH is set (dev compose)
-- All dev users have password: "password"
-- pw_hash is bcrypt cost-10 hash of "password"
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Persons (15 users with explicit IDs, mixed roles)
-- -----------------------------------------------------------------------------
INSERT INTO persons (person_id, person_name, person_email, username, pw_hash, role_id, locked) VALUES
    (1,  'Admin User',       'admin@factory.com',     'admin',      '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 77, FALSE),
    (5,  'Alice Martin',     'alice@factory.com',     'alice',      '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 20, FALSE),
    (10, 'Bob Dupont',       'bob@factory.com',       'bob',        '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 20, FALSE),
    (13, 'Carol Schmidt',    'carol@factory.com',     'carol',      '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 10, FALSE),
    (15, 'David Kovacs',     'david@factory.com',     'david',      '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 10, FALSE),
    (18, 'Eve Moreau',       'eve@factory.com',       'eve',        '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 10, FALSE),
    (20, 'Frank Weber',      'frank@factory.com',     'frank',      '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 1,  FALSE),
    (22, 'Grace Toth',       'grace@factory.com',     'grace',      '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 1,  FALSE),
    (25, 'Hans Bauer',       'hans@factory.com',      'hans',       '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 1,  FALSE),
    (27, 'Irene Lefevre',    'irene@factory.com',     'irene',      '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 10, FALSE),
    (29, 'Jan Novak',        'jan@factory.com',       'jan',        '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 20, FALSE),
    (30, 'Karl Muller',      'karl@factory.com',      'karl',       '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 1,  FALSE),
    (31, 'Laura Patel',      'laura@factory.com',     'laura',      '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 1,  FALSE),
    (32, 'Marco Rossi',      'marco@factory.com',     'marco',      '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 10, FALSE),
    (33, 'Nina Horvath',     'nina@factory.com',      'nina',       '$2a$10$Vi9M9nEBf0V9/hIdnX4N..HHfBcP9TYAvpknxBxAib62TsDwVi3Qm', 1,  FALSE)
ON CONFLICT DO NOTHING;
SELECT setval('persons_person_id_seq', GREATEST(33, (SELECT MAX(person_id) FROM persons)));

-- -----------------------------------------------------------------------------
-- Preferences
-- -----------------------------------------------------------------------------
INSERT INTO preferences (person_id, language, theme, active_columns) VALUES
    (29, 'en', FALSE, '{
        "id": true,
        "risk_type": true,
        "machines": true,
        "line_id": true,
        "category_id": true,
        "danger_description": true,
        "risk_level_id": true,
        "action": true,
        "responsible_id": true,
        "deadline": true,
        "action_state_id": true,
        "implementation_date": false,
        "cost": false,
        "expense_type_id": false,
        "moc_number": false,
        "created_by_id": false,
        "creation_date": false
    }'),
    (1,  'en', TRUE,  NULL),
    (5,  'en', FALSE, NULL),
    (10, 'en', FALSE, NULL),
    (13, 'en', FALSE, NULL)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- Risks (105 rows, inserted without linked_risk_id first)
-- Columns: id, risk_type, machines, line_id, category_id, danger_description,
--          risk_level_id, action, responsible_id, moc_number, deadline,
--          action_state_id, implementation_date, created_by_id, committed_at,
--          cost, expense_type_id
-- -----------------------------------------------------------------------------
INSERT INTO risks (id, risk_type, machines, line_id, category_id, danger_description, risk_level_id, action, responsible_id, moc_number, deadline, action_state_id, implementation_date, created_by_id, committed_at, cost, expense_type_id) VALUES
-- Risks 1-3: showcase entries (id=1,2,3 matching frontend examples)
(1,  1, 'Press Machine P-101', 1, 1, 'Unguarded rotating shaft poses entanglement risk to operators during routine maintenance.', 1, 'Install fixed guard with interlocked access panel', 15, NULL, '2026-04-15', 3, NULL, 13, '2026-01-10 09:00:00', 3500.00, 1),
(2,  2, 'Press Machine P-101', 1, 1, 'Long-term metal fatigue in main drive shaft could cause catastrophic failure.', 2, 'Schedule full shaft inspection and replacement programme', 13, 5001, '2026-09-30', 2, NULL, 13, '2026-01-10 09:15:00', 28000.00, 1),
(3,  1, 'Conveyor Belt CB-03', 2, 3, 'Gap between conveyor and fixed structure allows foot entrapment at transfer point.', 2, 'Install toe-board and anti-entrapment barrier', 15, NULL, '2026-03-31', 1, NULL, 5, NULL, 800.00, 2),

-- Risks 4-20: short-term risks, various states/levels/lines
(4,  1, 'Forklift FLT-02',      1, 4, 'No audible warning when reversing in congested zone.', 1, 'Fit reversing alarm and install mirror at blind corner', 18, NULL, '2026-03-20', 4, NULL, 5, '2026-01-12 10:00:00', 650.00, 2),
(5,  1, 'CNC Lathe L-07',       2, 2, 'Coolant leak onto floor creating slip hazard near electrical panel.', 2, 'Repair coolant hose and install drip tray', 15, NULL, '2026-03-25', 7, '2026-03-15', 13, '2026-01-15 08:30:00', 420.00, 2),
(6,  1, 'Robot Arm RA-04',      3, 5, 'Lockout/tagout procedure not posted at robot control station.', 1, 'Print and laminate LOTO procedure, mount at station', 27, NULL, '2026-03-18', 1, NULL, 10, NULL, 50.00, 2),
(7,  1, 'Overhead Crane OC-01', 4, 9, 'Crane hook latch worn, risk of load dropping.', 1, 'Replace hook assembly immediately', 13, NULL, '2026-03-15', 5, NULL, 5, '2026-01-20 11:00:00', 1200.00, 2),
(8,  1, 'Welding Station WS-02',4, 8, 'Fire extinguisher expired, station has no active suppression.', 2, 'Replace extinguisher and schedule monthly inspection', 18, NULL, '2026-03-22', 2, NULL, 29, '2026-02-01 09:00:00', 180.00, 2),
(9,  1, 'Pallet Jack PJ-05',    5, 7, 'Operator reports wrist strain due to non-ergonomic handle height.', 3, 'Replace with powered pallet truck', 32, NULL, '2026-05-01', 1, NULL, 13, NULL, 4200.00, 1),
(10, 1, 'Paint Booth PB-01',    6, 6, 'Ventilation fan belt broken, vapour accumulation risk.', 1, 'Replace fan belt and verify airflow before resuming operation', 15, NULL, '2026-03-16', 3, NULL, 10, '2026-02-05 14:00:00', 250.00, 2),
(11, 1, 'Hydraulic Press HP-02',1, 1, 'Safety interlock bypassed with tape, press can operate with guard open.', 1, 'Remove bypass and restore interlock, retrain operator', 13, NULL, '2026-03-14', 7, '2026-03-10', 5, '2026-01-25 08:00:00', 200.00, 2),
(12, 1, 'Grinding Wheel GW-03', 2, 1, 'Wheel flange cracked, risk of wheel burst.', 1, 'Stop use immediately, replace wheel and flange', 27, NULL, '2026-03-13', 7, '2026-03-12', 13, '2026-02-10 10:00:00', 320.00, 2),
(13, 1, 'Assembly Jig AJ-01',   4, 9, 'Jig bolts loose, risk of workpiece ejection during clamping.', 2, 'Tighten bolts and implement daily torque check', 15, NULL, '2026-04-01', 2, NULL, 29, '2026-02-12 09:30:00', 0.00, 2),
(14, 1, 'Scissor Lift SL-01',   5, 9, 'Fall arrest harness attachment point missing on platform.', 1, 'Weld anchor point and certify before use', 18, NULL, '2026-03-28', 3, NULL, 5, '2026-02-14 11:00:00', 750.00, 2),
(15, 1, 'Compressor C-03',      6, 2, 'Pressure relief valve has never been tested.', 2, 'Schedule PRV test with certified engineer', 32, NULL, '2026-04-10', 1, NULL, 13, NULL, 300.00, 2),
(16, 1, 'Band Saw BS-01',       3, 1, 'Blade guard missing on lower blade section.', 1, 'Procure and fit lower blade guard', 27, NULL, '2026-03-26', 2, NULL, 10, '2026-02-20 08:00:00', 480.00, 2),
(17, 1, 'Chemical Store CS-01', 6, 6, 'Incompatible chemicals stored adjacent (acids + bases).', 1, 'Segregate storage and label cabinets', 15, NULL, '2026-03-19', 7, '2026-03-11', 5, '2026-02-18 13:00:00', 150.00, 2),
(18, 1, 'Drill Press DP-04',    1, 2, 'Damaged power cable, insulation exposed near chuck.', 1, 'Take out of service, replace cable before reuse', 13, NULL, '2026-03-15', 7, '2026-03-14', 13, '2026-02-22 10:00:00', 80.00, 2),
(19, 1, 'Loading Dock LD-02',   5, 3, 'Dock leveller not secured when truck departs, fall risk.', 2, 'Install dock lock system and train drivers', 18, NULL, '2026-04-20', 1, NULL, 29, NULL, 2800.00, 1),
(20, 1, 'Mixing Tank MT-01',    6, 6, 'No emergency stop accessible from operator position.', 1, 'Install emergency stop button within operator reach', 32, NULL, '2026-03-30', 3, NULL, 10, '2026-02-25 09:00:00', 600.00, 2),

-- Risks 21-40: long-term risks
(21, 2, 'Press Machine P-101',  1, 1, 'Ageing hydraulic system showing progressive seal degradation over 5 years.', 2, 'Full hydraulic system overhaul in next maintenance window', 13, 5002, '2026-12-31', 1, NULL, 5, NULL, 45000.00, 1),
(22, 2, 'Production Line A',    1, 2, 'Outdated PLC controllers approaching end of support life.', 3, 'Plan PLC replacement project with IT and OT teams', 5, 5003, '2027-06-30', 1, NULL, 13, NULL, 120000.00, 1),
(23, 2, 'Conveyor Belt CB-03',  2, 1, 'Frame shows fatigue cracks after 12 years of operation.', 2, 'Commission structural survey and schedule replacement', 10, 5004, '2026-10-31', 2, NULL, 5, '2026-01-20 10:00:00', 35000.00, 1),
(24, 2, 'Forklift FLT-02',      1, 9, 'Fleet age averaging 11 years, increasing maintenance burden.', 3, 'Develop fleet renewal plan over 3 years', 29, NULL, '2027-12-31', 1, NULL, 10, NULL, 180000.00, 1),
(25, 2, 'Ventilation System',   3, 6, 'HVAC system undersized for current production volume.', 2, 'Engineering study for HVAC upgrade', 13, 5005, '2026-11-30', 3, NULL, 5, '2026-01-25 14:00:00', 75000.00, 1),
(26, 2, 'Electrical Distribution',2,2, 'Main switchboard from 1998, no arc-flash protection.', 1, 'Full electrical audit and switchboard replacement', 10, 5006, '2026-08-31', 2, NULL, 13, '2026-02-01 09:00:00', 95000.00, 1),
(27, 2, 'Storage Racking SR-01',4, 9, 'Racking not rated for current maximum load per bay.', 2, 'Structural recalculation and load-limit signage programme', 5, NULL, '2026-09-30', 1, NULL, 10, NULL, 8000.00, 1),
(28, 2, 'Dust Extraction DE-01',3, 6, 'Combustible dust accumulation in ducting exceeds safe level.', 1, 'Deep clean ducting and implement quarterly inspection', 27, NULL, '2026-04-30', 3, NULL, 13, '2026-02-10 11:00:00', 12000.00, 2),
(29, 2, 'Compressed Air System',5, 2, 'Air dryers failing, moisture entering tools causing corrosion.', 3, 'Replace air dryers in phased programme', 32, 5007, '2026-07-31', 1, NULL, 5, NULL, 22000.00, 1),
(30, 2, 'Emergency Exits',      6, 4, 'Two emergency exits blocked by racking during peak periods.', 1, 'Relocate racking and install permanent exclusion zone markings', 18, NULL, '2026-04-15', 2, NULL, 29, '2026-02-15 08:30:00', 500.00, 2),
(31, 2, 'Noise Exposure',       1, 7, 'Ambient noise level 87 dB, above action level, no hearing conservation programme.', 2, 'Noise survey, implement hearing protection zone and programme', 13, NULL, '2026-06-30', 1, NULL, 10, NULL, 6000.00, 2),
(32, 2, 'Lighting System',      2, 2, 'LED replacement project stalled, 40% of fixtures still fluorescent.', 3, 'Complete LED retrofit within 12 months', 5, NULL, '2026-12-31', 1, NULL, 5, NULL, 28000.00, 1),
(33, 2, 'Chemical Inventory',   6, 6, 'SDS database not updated since 2022, >200 chemicals unreviewed.', 2, 'Full chemical inventory and SDS review project', 15, NULL, '2026-06-30', 2, NULL, 13, '2026-02-20 09:00:00', 3000.00, 2),
(34, 2, 'Training Records',     4, 9, 'No centralised training matrix, compliance gaps unknown.', 3, 'Implement digital training management system', 29, NULL, '2026-09-30', 1, NULL, 10, NULL, 15000.00, 1),
(35, 2, 'Fire Detection System',1, 8, 'Fire detection system not integrated with suppression in paint shop.', 1, 'System integration project with fire contractor', 13, 5008, '2026-07-31', 3, NULL, 5, '2026-02-25 10:00:00', 42000.00, 1),
(36, 2, 'Walkway Surfaces',     5, 3, 'Anti-slip coating worn on elevated walkways, slip risk.', 2, 'Recoat all elevated walkways during summer shutdown', 18, NULL, '2026-08-15', 1, NULL, 29, NULL, 9500.00, 2),
(37, 2, 'Machine Guards',       3, 1, 'Systematic guard audit identified 23 non-compliant machines.', 2, 'Phased guard upgrade programme over 18 months', 27, 5009, '2027-03-31', 2, NULL, 5, '2026-03-01 08:00:00', 38000.00, 1),
(38, 2, 'Permit to Work System',2, 9, 'Paper-based PTW system causing delays and compliance gaps.', 3, 'Implement digital permit-to-work system', 10, NULL, '2026-10-31', 1, NULL, 13, NULL, 18000.00, 1),
(39, 2, 'CCTV Coverage',        6, 9, 'CCTV blind spots in loading area, near-misses unrecorded.', 3, 'Install 4 additional cameras to eliminate blind spots', 5, NULL, '2026-05-31', 3, NULL, 29, '2026-03-05 09:00:00', 5500.00, 2),
(40, 2, 'Machinery Vibration',  4, 7, 'HAV exposure exceeds EAV for 3 operators on grinding tasks.', 2, 'Source low-vibration tool alternatives and implement rotation', 13, NULL, '2026-08-31', 1, NULL, 5, NULL, 7000.00, 2),

-- Risks 41-60: mix with various states/deadlines
(41, 1, 'Pump Station PS-01',   1, 2, 'Pump overheating due to blocked impeller, burn risk.', 2, 'Clean impeller and install temperature alarm', 15, NULL, '2026-04-05', 2, NULL, 13, '2026-02-28 10:00:00', 900.00, 2),
(42, 1, 'Access Ladder AL-02',  4, 3, 'Ladder rung bent, slip risk when ascending to mezzanine.', 2, 'Replace ladder rung or full ladder replacement', 27, NULL, '2026-03-28', 3, NULL, 10, '2026-03-02 09:00:00', 350.00, 2),
(43, 1, 'Spray Booth SB-01',    6, 8, 'Spray booth fire damper stuck open, fire barrier compromised.', 1, 'Service damper and test actuator', 32, NULL, '2026-03-22', 4, NULL, 5, '2026-03-03 08:30:00', 480.00, 2),
(44, 1, 'Packing Machine PM-02',6, 1, 'In-running nip point on packing machine exposed during operation.', 1, 'Fit fixed barrier guard at nip point', 18, NULL, '2026-04-08', 1, NULL, 29, NULL, 700.00, 2),
(45, 1, 'Air Compressor AC-02', 5, 2, 'Compressor safety valve corroded, may not open at set pressure.', 1, 'Replace safety valve immediately', 13, NULL, '2026-03-17', 7, '2026-03-16', 13, '2026-03-04 10:00:00', 220.00, 2),
(46, 1, 'Trolley TR-05',        2, 7, 'Trolley handle broken, operator using improvised grip.', 3, 'Replace trolley handle', 15, NULL, '2026-04-15', 1, NULL, 15, NULL, 85.00, 2),
(47, 1, 'Milling Machine MM-01',3, 1, 'Coolant pipe cracked, coolant spraying toward electrical panel.', 2, 'Replace coolant pipe and install splash guard', 27, NULL, '2026-03-24', 3, NULL, 10, '2026-03-05 09:00:00', 380.00, 2),
(48, 1, 'Packaging Line PL-01', 6, 1, 'Rejected product diverter jam causes operator to reach into machine.', 2, 'Redesign diverter with auto-clear function', 32, NULL, '2026-05-15', 2, NULL, 5, '2026-03-06 08:00:00', 2800.00, 2),
(49, 1, 'Loading Bay LB-03',    5, 3, 'External loading bay has no edge protection when dock-leveller raised.', 2, 'Install retractable edge barrier', 18, NULL, '2026-04-30', 1, NULL, 29, NULL, 1800.00, 2),
(50, 1, 'Steam Trap ST-01',     1, 9, 'Steam trap blowing live steam, burn risk to passing workers.', 1, 'Isolate and replace steam trap immediately', 13, NULL, '2026-03-16', 7, '2026-03-15', 5, '2026-03-07 10:00:00', 560.00, 2),
(51, 1, 'Filling Machine FM-01',6, 6, 'Solvent filling machine lacks adequate LEV, vapour accumulation.', 2, 'Install dedicated LEV at filling point', 15, NULL, '2026-04-20', 2, NULL, 13, '2026-03-08 09:30:00', 3200.00, 2),
(52, 1, 'Hand Truck HT-07',     4, 7, 'Worker reports lower back pain from overloading hand truck.', 3, 'Set load limit 25 kg and provide manual handling training', 27, NULL, '2026-05-01', 1, NULL, 10, NULL, 200.00, 2),
(53, 1, 'Roller Conveyor RC-04',2, 1, 'Nip point between rollers and frame at inload end.', 1, 'Install nip-point guard at inload end', 32, NULL, '2026-04-10', 3, NULL, 5, '2026-03-09 08:00:00', 650.00, 2),
(54, 1, 'Tank Farm TF-01',      6, 6, 'Bund drainage valve left open, spill containment ineffective.', 2, 'Lock drainage valve and implement inspection checklist', 18, NULL, '2026-03-20', 5, NULL, 29, '2026-03-10 10:00:00', 0.00, 2),
(55, 1, 'Platform PL-02',       3, 3, 'Handrail on maintenance platform loose, fall from height risk.', 1, 'Re-weld handrail socket before platform use', 13, NULL, '2026-03-19', 3, NULL, 13, '2026-03-10 14:00:00', 280.00, 2),
(56, 1, 'Power Tool PT-03',     1, 2, 'Angle grinder used without face shield by operator.', 3, 'Issue PPE and enforce face shield policy', 15, NULL, '2026-03-25', 7, '2026-03-18', 15, '2026-03-11 08:00:00', 120.00, 2),
(57, 1, 'Palletiser PA-01',     6, 1, 'Palletiser safety fence gate can be opened during cycle.', 1, 'Repair gate interlock switch', 27, NULL, '2026-04-02', 2, NULL, 10, '2026-03-11 09:00:00', 380.00, 2),
(58, 1, 'Inspection Pit IP-01', 4, 3, 'Inspection pit open without barriers when not in use.', 2, 'Install retractable pit cover or permanent barrier', 32, NULL, '2026-04-25', 1, NULL, 5, NULL, 2200.00, 1),
(59, 1, 'Gas Cylinder Store GS',3, 6, 'Gas cylinders not chained, tip-over risk.', 2, 'Install cylinder restraint chains', 18, NULL, '2026-03-21', 3, NULL, 29, '2026-03-12 10:00:00', 180.00, 2),
(60, 1, 'Roof Access RA-01',    1, 3, 'No safe means of access to roof for maintenance personnel.', 2, 'Install permanent access ladder with gate and anchor points', 13, NULL, '2026-06-30', 2, NULL, 13, '2026-03-12 11:00:00', 8500.00, 1),

-- Risks 61-80: more variety, some with cost/expense, some with moc
(61, 1, 'Strapping Machine SM-01',6,1,'Retractable knife not retracting after strapping cycle.', 1, 'Service knife mechanism and test before return to service', 15, NULL, '2026-03-23', 7, '2026-03-20', 15, '2026-01-15 10:00:00', 320.00, 2),
(62, 2, 'Boiler Room BR-01',    1, 8, 'Boiler inspection overdue by 6 months.', 1, 'Schedule statutory boiler inspection immediately', 13, NULL, '2026-04-30', 3, NULL, 5, '2026-01-20 09:00:00', 1800.00, 2),
(63, 1, 'Assembly Station AS-04',4,7,'Monitor positioned too low, operators working with neck bent.', 3, 'Install adjustable monitor arm', 27, NULL, '2026-05-01', 1, NULL, 10, NULL, 350.00, 2),
(64, 2, 'Roof Structure',       1, 9, 'Roof cladding showing corrosion, structural survey not done.', 2, 'Commission roof structural survey', 5, 5010, '2026-09-30', 1, NULL, 13, NULL, 5000.00, 2),
(65, 1, 'Injection Moulding IM-01',3,1,'Safety door switch worn, machine starts with guard open.', 1, 'Replace door switch immediately', 32, NULL, '2026-03-18', 7, '2026-03-17', 13, '2026-01-22 08:00:00', 280.00, 2),
(66, 2, 'Sprinkler System',     2, 8, 'Sprinkler heads in warehouse blocked by racking above 3 m level.', 2, 'Reorganise racking to maintain 0.5 m clear zone below sprinklers', 10, 5011, '2026-06-30', 2, NULL, 5, '2026-01-25 10:00:00', 0.00, 2),
(67, 1, 'Valve Station VS-02',  5, 2, 'Valve station lacks isolation tags, risk of inadvertent operation.', 2, 'Label all valves with colour-coded tags and position indicators', 18, NULL, '2026-04-12', 3, NULL, 29, '2026-01-28 09:00:00', 420.00, 2),
(68, 2, 'Safety Shower',        6, 6, 'Chemical area safety shower water temperature too cold in winter.', 3, 'Install tempered water supply to safety shower', 15, NULL, '2026-10-31', 1, NULL, 13, NULL, 4500.00, 2),
(69, 1, 'Grinding Machine GM-02',2,1,'Grinding rest gap too large, workpiece can be drawn between wheel and rest.', 1, 'Adjust rest gap to <3 mm per standard', 27, NULL, '2026-03-22', 7, '2026-03-20', 10, '2026-01-30 10:00:00', 0.00, 2),
(70, 2, 'Site Drainage',        5, 9, 'Stormwater drain routes through process area, contamination risk.', 3, 'Design bunded drainage separation project', 5, 5012, '2027-06-30', 1, NULL, 10, NULL, 65000.00, 1),
(71, 1, 'Ladder Storage LS-01', 4, 3, 'Aluminium ladders stored horizontally, trip hazard.', 3, 'Install vertical ladder rack', 32, NULL, '2026-04-18', 2, NULL, 5, '2026-02-02 09:00:00', 250.00, 2),
(72, 2, 'Pedestrian Routes',    3, 9, 'Pedestrian and forklift routes share 3 corridor sections.', 2, 'Mark dedicated pedestrian corridors and install barriers', 13, NULL, '2026-07-31', 3, NULL, 13, '2026-02-05 10:00:00', 6500.00, 2),
(73, 1, 'Roller Shutter RS-01', 1, 2, 'Roller shutter safety edge sensor disconnected.', 1, 'Reconnect and test safety edge before shutter use', 15, NULL, '2026-03-17', 7, '2026-03-16', 15, '2026-02-08 08:00:00', 180.00, 2),
(74, 2, 'Eyewash Station',      6, 6, 'Eyewash station inspection not recorded for 4 months.', 3, 'Implement weekly eyewash inspection log', 27, NULL, '2026-04-30', 1, NULL, 5, NULL, 50.00, 2),
(75, 1, 'Sheet Metal Store SMS',4, 1, 'Sharp sheet metal edges exposed in storage area, laceration risk.', 2, 'Install edge protectors on stored sheet metal', 32, NULL, '2026-03-27', 3, NULL, 29, '2026-02-10 11:00:00', 320.00, 2),
(76, 2, 'Security Fencing',     2, 9, 'Perimeter fencing gap allows unauthorised pedestrian access to plant.', 3, 'Repair fencing and add motion-activated lighting', 10, NULL, '2026-06-30', 1, NULL, 13, NULL, 3500.00, 2),
(77, 1, 'Vacuum Pump VP-01',    5, 2, 'Vacuum pump emitting unusual noise, bearing failure imminent.', 2, 'Replace bearings during next planned shutdown', 18, NULL, '2026-04-05', 2, NULL, 29, '2026-02-14 09:30:00', 1100.00, 2),
(78, 2, 'Arc Flash Study',      1, 2, 'No arc flash hazard analysis done for new switchgear installed 2024.', 1, 'Engage electrical engineer for arc flash study', 5, 5013, '2026-08-31', 2, NULL, 5, '2026-02-18 10:00:00', 12000.00, 2),
(79, 1, 'Hoist H-03',           3, 1, 'Hoist brake slipping under rated load.', 1, 'Take out of service, overhaul brake assembly', 13, NULL, '2026-03-19', 3, NULL, 13, '2026-02-20 08:00:00', 1600.00, 2),
(80, 2, 'Waste Management',     6, 6, 'Hazardous waste segregation not enforced, mixed waste going to landfill.', 2, 'Implement waste segregation programme with contractor', 15, NULL, '2026-06-30', 1, NULL, 10, NULL, 4000.00, 2),

-- Risks 81-105: final batch
(81,  1, 'Cable Tray CT-01',    2, 2, 'Cable tray overloaded, cables sagging below 2 m head height.', 2, 'Install additional cable tray support brackets', 27, NULL, '2026-04-08', 2, NULL, 5, '2026-02-22 09:00:00', 750.00, 2),
(82,  2, 'Heat Stress',         1, 7, 'Foundry area temperatures exceed 28°C WBGT in summer months.', 2, 'Install evaporative coolers and implement heat stress monitoring', 10, NULL, '2026-05-31', 1, NULL, 13, NULL, 9500.00, 2),
(83,  1, 'Pressure Gauge PG-04',5, 2, 'Pressure gauge on air receiver showing no reading, instrument failure.', 2, 'Replace pressure gauge before system use', 32, NULL, '2026-03-18', 7, '2026-03-17', 13, '2026-02-25 10:00:00', 95.00, 2),
(84,  2, 'Contractor Management',4,9,'No site-specific induction for contractors, safety gaps.', 2, 'Develop contractor management procedure and e-learning induction', 5, NULL, '2026-07-31', 1, NULL, 5, NULL, 3000.00, 2),
(85,  1, 'Scrap Bin SB-03',     3, 1, 'Overfilled scrap bin, metal pieces overhanging, drop risk.', 2, 'Empty bin and set fill-level limit mark', 18, NULL, '2026-03-24', 7, '2026-03-22', 18, '2026-03-01 08:00:00', 0.00, 2),
(86,  2, 'Safety Signage',      2, 9, 'Safety signs faded and non-compliant with ISO 7010 standard.', 3, 'Audit and replace all safety signage', 13, NULL, '2026-09-30', 1, NULL, 13, NULL, 2200.00, 2),
(87,  1, 'Pipework PW-07',      6, 6, 'Chemical pipe flange leaking, minor drip creating corrosion.', 2, 'Replace flange gasket and inspect neighbouring joints', 15, NULL, '2026-03-25', 3, NULL, 15, '2026-03-03 09:30:00', 420.00, 2),
(88,  2, 'First Aid Kits',      5, 4, 'First aid kits depleted and not replenished after incidents.', 3, 'Implement monthly inspection and replenishment procedure', 27, NULL, '2026-04-30', 1, NULL, 10, NULL, 800.00, 2),
(89,  1, 'Stretcher ST-02',     4, 4, 'Rescue stretcher not accessible — stored behind locked cabinet without key nearby.', 2, 'Relocate stretcher to accessible location with posted key', 32, NULL, '2026-03-23', 7, '2026-03-22', 5, '2026-03-05 10:00:00', 0.00, 2),
(90,  2, 'PPE Inspection',      1, 9, 'No formal PPE inspection and replacement programme.', 3, 'Implement PPE register with annual inspection cycle', 10, NULL, '2026-06-30', 1, NULL, 13, NULL, 1500.00, 2),
(91,  1, 'Chain Block CB-04',   3, 1, 'Chain block chain kinked, risk of failure under load.', 1, 'Take out of service and replace chain', 13, NULL, '2026-03-16', 7, '2026-03-15', 13, '2026-03-07 08:00:00', 480.00, 2),
(92,  2, 'Safety Culture',      2, 9, 'Near-miss reporting rate <10% of estimated incidents.', 3, 'Launch near-miss reporting campaign and anonymous reporting system', 5, NULL, '2026-09-30', 1, NULL, 5, NULL, 5000.00, 2),
(93,  1, 'Tank Level Sensor',   6, 2, 'Level sensor failed, operator manually checking tank with torch.', 2, 'Replace level sensor', 15, NULL, '2026-04-03', 2, NULL, 15, '2026-03-08 09:00:00', 780.00, 2),
(94,  2, 'Machinery Guarding Strategy',3,1,'No site-wide machine guarding standard, inconsistent implementation.', 2, 'Develop and publish site machine guarding standard', 27, NULL, '2026-10-31', 1, NULL, 10, NULL, 4000.00, 2),
(95,  1, 'Torque Wrench TW-02', 4, 9, 'Torque wrench calibration expired 3 months ago.', 3, 'Send for calibration before next use', 32, NULL, '2026-03-20', 7, '2026-03-18', 5, '2026-03-10 10:00:00', 150.00, 2),
(96,  2, 'Explosion Protection',6, 6, 'ATEX zone classification review overdue by 2 years.', 1, 'Commission ATEX review by competent person', 18, 5014, '2026-06-30', 2, NULL, 13, '2026-03-11 09:00:00', 8000.00, 2),
(97,  1, 'Fork Pocket FP-01',   5, 1, 'Fork pocket on heavy racking damaged by forklift impact.', 1, 'Take racking bay out of use, replace damaged upright', 13, NULL, '2026-03-17', 3, NULL, 13, '2026-03-11 10:00:00', 1400.00, 2),
(98,  2, 'Spill Response',      1, 6, 'Insufficient spill kit capacity for largest chemical container on site.', 2, 'Increase spill kit capacity and train response team', 5, NULL, '2026-05-31', 1, NULL, 5, NULL, 2500.00, 2),
(99,  1, 'Flexible Hose FH-03', 2, 6, 'Flexible chemical hose shows external deterioration, bulging visible.', 1, 'Replace hose immediately, inspect adjacent hoses', 10, NULL, '2026-03-15', 7, '2026-03-14', 10, '2026-03-12 08:00:00', 380.00, 2),
(100, 2, 'Risk Assessment Reviews',4,9,'Majority of risk assessments >3 years old, not reviewed after process changes.', 2, 'Systematic review programme targeting highest-risk tasks first', 29, NULL, '2026-12-31', 1, NULL, 29, NULL, 6000.00, 2),
(101, 1, 'Palletising Robot PR-01',6,1,'Safety scanner reflector dirty, reducing effective range of protection.', 2, 'Clean all scanner reflectors and test detection zone', 15, NULL, '2026-03-26', 3, NULL, 15, '2026-03-12 09:00:00', 0.00, 2),
(102, 2, 'Subcontractor Electrical',2,2,'Site does not verify electrical competence of subcontractor engineers.', 2, 'Implement competency verification process for subcontractor electrical work', 27, NULL, '2026-08-31', 1, NULL, 10, NULL, 2000.00, 2),
(103, 1, 'Pipe Lagging PL-03',  1, 9, 'Steam pipe lagging missing on 4 m section, surface temperature >80°C.', 2, 'Re-lag steam pipe section before winter shutdown', 32, NULL, '2026-04-15', 2, NULL, 5, '2026-03-12 10:00:00', 900.00, 2),
(104, 2, 'Emergency Lighting',  5, 4, 'Emergency lighting test log shows 30% of fittings not tested for >12 months.', 2, 'Test all emergency lighting and repair/replace failed units', 18, NULL, '2026-05-31', 1, NULL, 29, NULL, 3500.00, 2),
(105, 1, 'Maintenance Scaffold', 3,3, 'Scaffold erected by operations staff without competency check.', 1, 'Dismantle and re-erect by competent scaffolder', 13, NULL, '2026-03-21', 7, '2026-03-20', 13, '2026-03-12 11:00:00', 800.00, 2)
ON CONFLICT DO NOTHING;

SELECT setval('risks_id_seq', GREATEST(105, (SELECT MAX(id) FROM risks)));

-- -----------------------------------------------------------------------------
-- Link paired risks (short-term ↔ long-term, must be opposite types)
-- Pairs: (1,2), (3,23), (10,25), (11,35), (26,73), (28,59)
-- -----------------------------------------------------------------------------
UPDATE risks SET linked_risk_id = 2   WHERE id = 1;
UPDATE risks SET linked_risk_id = 1   WHERE id = 2;
UPDATE risks SET linked_risk_id = 23  WHERE id = 3;
UPDATE risks SET linked_risk_id = 3   WHERE id = 23;
UPDATE risks SET linked_risk_id = 25  WHERE id = 10;
UPDATE risks SET linked_risk_id = 10  WHERE id = 25;
UPDATE risks SET linked_risk_id = 35  WHERE id = 11;
UPDATE risks SET linked_risk_id = 11  WHERE id = 35;
UPDATE risks SET linked_risk_id = 73  WHERE id = 26;
UPDATE risks SET linked_risk_id = 26  WHERE id = 73;

-- -----------------------------------------------------------------------------
-- Comments (15 rows)
-- -----------------------------------------------------------------------------
INSERT INTO comments (id, risk_id, author_id, content, created_at) VALUES
    (1,  1,  13, 'Guard design reviewed with supplier, installation scheduled for next week.', '2026-01-15 10:30:00'),
    (2,  1,  15, 'Installation confirmed complete, awaiting sign-off inspection.', '2026-02-03 14:00:00'),
    (3,  2,  5,  'MOC approved by engineering committee. Shaft replacement budgeted in Q3 project list.', '2026-01-22 09:15:00'),
    (4,  3,  5,  'Toe-board materials on order, ETA 2 weeks.', '2026-01-18 11:00:00'),
    (5,  7,  13, 'New hook assembly received. Crane out of service until installation.', '2026-01-21 15:00:00'),
    (6,  10, 15, 'Fan belt replaced at 14:30, airflow tested and confirmed adequate. Back in service.', '2026-02-06 16:00:00'),
    (7,  11, 13, 'Operator retraining completed. Interlock confirmed operational by maintenance.', '2026-01-26 10:00:00'),
    (8,  21, 13, 'Hydraulic survey quotations requested from 3 contractors.', '2026-02-05 09:00:00'),
    (9,  21, 5,  'Preferred contractor selected. Work scheduled for summer shutdown week 32.', '2026-03-01 14:00:00'),
    (10, 26, 10, 'Arc flash study linked to this risk (risk ID 78). Both being managed together.', '2026-02-02 10:30:00'),
    (11, 37, 5,  'Guard audit complete. 23 machines non-compliant, prioritised by risk level.', '2026-03-02 09:30:00'),
    (12, 37, 27, 'First 6 machines upgraded. Remaining 17 in Q2-Q3 programme.', '2026-03-12 11:00:00'),
    (13, 54, 29, 'Drainage valve now locked closed. Key held by shift supervisor only.', '2026-03-11 09:00:00'),
    (14, 72, 13, 'Pedestrian route marking design approved. Contractor starting next Monday.', '2026-03-06 10:00:00'),
    (15, 96, 18, 'ATEX consultant engaged. Site visit scheduled for April.', '2026-03-12 14:00:00')
ON CONFLICT DO NOTHING;
SELECT setval('comments_id_seq', GREATEST(15, (SELECT MAX(id) FROM comments)));

-- -----------------------------------------------------------------------------
-- Status History (25 rows — tracks state transitions for a subset of risks)
-- -----------------------------------------------------------------------------
INSERT INTO status_history (id, risk_id, old_status, new_status, reason, changed_by, changed_at) VALUES
    (1,  1,  1, 2, 'Assigned to responsible engineer.',                              13, '2026-01-11 09:00:00'),
    (2,  1,  2, 3, 'Guard procurement and installation work started.',               15, '2026-01-20 10:00:00'),
    (3,  2,  1, 2, 'Assigned after MOC approval.',                                   5,  '2026-01-22 09:30:00'),
    (4,  5,  1, 7, 'Coolant hose repaired and drip tray installed, risk resolved.',  15, '2026-03-15 16:00:00'),
    (5,  7,  1, 5, 'New hook arrived but incorrect part number ordered.',             13, '2026-01-22 09:00:00'),
    (6,  10, 1, 2, 'Assigned to mechanical maintenance.',                            15, '2026-02-06 08:00:00'),
    (7,  10, 2, 3, 'Belt replacement in progress.',                                  15, '2026-02-06 13:00:00'),
    (8,  10, 3, 7, 'Belt replaced and flow verified. Closed.',                       15, '2026-02-06 16:30:00'),
    (9,  11, 1, 3, 'Immediate rectification and operator retraining.',               13, '2026-01-25 08:30:00'),
    (10, 11, 3, 7, 'Interlock restored and tested. Operator retrained.',             13, '2026-01-26 11:00:00'),
    (11, 12, 1, 7, 'Wheel and flange replaced. Back in service.',                    27, '2026-02-12 14:00:00'),
    (12, 17, 1, 7, 'Chemicals segregated, labels applied.',                          15, '2026-03-11 12:00:00'),
    (13, 18, 1, 7, 'Cable replaced by electrician, PAT tested.',                     13, '2026-03-14 15:00:00'),
    (14, 23, 1, 2, 'Structural survey company appointed.',                           10, '2026-01-21 10:00:00'),
    (15, 26, 1, 2, 'Electrical audit contractor selected.',                          10, '2026-02-02 09:00:00'),
    (16, 30, 1, 2, 'Racking relocation plan drafted.',                               18, '2026-02-16 10:00:00'),
    (17, 37, 1, 2, 'Guard audit complete, prioritisation list agreed.',              5,  '2026-03-02 10:00:00'),
    (18, 43, 1, 4, 'Damper serviced, awaiting third-party verification.',            32, '2026-03-04 15:00:00'),
    (19, 45, 1, 7, 'Safety valve replaced by certified engineer.',                   13, '2026-03-16 14:00:00'),
    (20, 50, 1, 7, 'Steam trap isolated and replaced. Steam-free after repair.',     5,  '2026-03-15 11:00:00'),
    (21, 54, 1, 5, 'Root cause: lack of awareness. Requires procedure update.',      29, '2026-03-11 09:30:00'),
    (22, 56, 1, 7, 'PPE issued and face shield policy communicated at toolbox talk.', 15, '2026-03-18 10:00:00'),
    (23, 65, 1, 7, 'Door switch replaced and tested by electrician.',                13, '2026-03-17 15:00:00'),
    (24, 69, 1, 7, 'Grinding rest gap adjusted and measured. Within tolerance.',     10, '2026-03-20 09:30:00'),
    (25, 79, 1, 3, 'Hoist taken out of service, brake overhaul started.',            13, '2026-02-21 08:00:00')
ON CONFLICT DO NOTHING;
SELECT setval('status_history_id_seq', GREATEST(25, (SELECT MAX(id) FROM status_history)));

-- -----------------------------------------------------------------------------
-- Deadline History (7 rows)
-- -----------------------------------------------------------------------------
INSERT INTO deadline_history (id, risk_id, old_deadline, new_deadline, reason, changed_by, changed_at) VALUES
    (1, 2,  '2026-06-30', '2026-09-30', 'Engineering resource not available until Q3. MOC timeline extended.', 5,  '2026-02-15 10:00:00'),
    (2, 9,  '2026-04-01', '2026-05-01', 'Powered pallet truck on back-order, delivery delayed 4 weeks.',       32, '2026-03-20 09:00:00'),
    (3, 21, '2026-09-30', '2026-12-31', 'Hydraulic overhaul scope increased after further inspection.',         13, '2026-02-28 11:00:00'),
    (4, 25, '2026-09-30', '2026-11-30', 'Engineering study delayed by consultant availability.',               13, '2026-02-20 09:30:00'),
    (5, 35, '2026-06-30', '2026-07-31', 'Fire contractor schedule conflict, rescheduled.',                     5,  '2026-03-05 10:00:00'),
    (6, 48, '2026-04-30', '2026-05-15', 'Design rework needed after prototype testing revealed gap.',          32, '2026-03-08 14:00:00'),
    (7, 60, '2026-05-31', '2026-06-30', 'Planning permission application delayed by one month.',               13, '2026-03-10 09:00:00')
ON CONFLICT DO NOTHING;
SELECT setval('deadline_history_id_seq', GREATEST(7, (SELECT MAX(id) FROM deadline_history)));
