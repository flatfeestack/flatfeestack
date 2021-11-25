-- Create Sponsors
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('7a760bac-5d84-498b-9757-5c913d35c605', NULL, NULL, NULL, NULL, NULL, NULL, 'sponsor1@flatfeestack.io', NULL, NULL, NULL, '6IREUFR7ST3UENSNKOBOWMSIXMMXU===', 'USR', '2021-10-30 12:00:53.293937');
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('67d62066-b966-425d-8060-2a58a17b48c9', NULL, NULL, NULL, NULL, NULL, null, 'sponsor2@flatfeestack.io', NULL, NULL, NULL, 'BW56QCVUHPPTCM5DMAT7K774X4PDA===', 'USR', '2021-10-30 11:59:02.935403');
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('23fbded4-6d6b-446e-8b33-9cefcacc9a01', NULL, NULL, NULL, NULL, NULL, null, 'sponsor3@flatfeestack.io', NULL, NULL, NULL, 'CD2AZO2S7Y73MR76Y6PLKPIYAF3QU===', 'USR', '2021-10-30 11:59:42.816161');
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('d994260b-125e-441a-926a-bd0498bfb902', NULL, NULL, NULL, NULL, NULL, null, 'sponsor4@flatfeestack.io', NULL, NULL, NULL, 'YQQYQODRNSLDMONVEDLMUVE34T4AC===', 'USR', '2021-10-30 12:00:12.234188');
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('fdc85b12-3fdf-11ec-9356-0242ac130003', NULL, NULL, NULL, NULL, NULL, null, 'sponsor5@flatfeestack.io', NULL, NULL, NULL, 'YQQYQODRNSLDMONVEDLMUVE34T4AC===', 'USR', '2021-10-30 12:00:12.234188');

-- Create Contributor
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('d263a9f0-3fde-11ec-9356-0242ac130003', NULL, NULL, NULL, NULL, NULL, NULL, 'contributor1@flatfeestack.io', NULL, NULL, NULL, '6IREUFR7ST3UENSNKOBOWMSIXMMXU===', 'USR', '2021-10-30 12:00:53.293937');
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('14b8f83c-3fdf-11ec-9356-0242ac130003', NULL, NULL, NULL, NULL, NULL, null, 'contributor2@flatfeestack.io', NULL, NULL, NULL, 'BW56QCVUHPPTCM5DMAT7K774X4PDA===', 'USR', '2021-10-30 11:59:02.935403');
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('18a2da08-3fdf-11ec-9356-0242ac130003', NULL, NULL, NULL, NULL, NULL, null, 'contributor3@flatfeestack.io', NULL, NULL, NULL, 'CD2AZO2S7Y73MR76Y6PLKPIYAF3QU===', 'USR', '2021-10-30 11:59:42.816161');
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('1caeda7a-3fdf-11ec-9356-0242ac130003', NULL, NULL, NULL, NULL, NULL, null, 'contributor4@flatfeestack.io', NULL, NULL, NULL, 'YQQYQODRNSLDMONVEDLMUVE34T4AC===', 'USR', '2021-10-30 12:00:12.234188');

-- Git Email for Contributors
INSERT INTO public.git_email (id, user_id, email, token, confirmed_at, created_at) VALUES ('6a27baf5-3ed4-41f7-8317-94d844ac75b2', 'd263a9f0-3fde-11ec-9356-0242ac130003', 'contributor1@flatfeestack.io', 'C64IVABILT4SRAROBBWDNMVY2E======', NULL, '2021-10-30 12:48:59.413386');
INSERT INTO public.git_email (id, user_id, email, token, confirmed_at, created_at) VALUES ('08d3a39c-5b41-44bb-b7f6-0995a0d51c67', '14b8f83c-3fdf-11ec-9356-0242ac130003', 'contributor2@flatfeestack.io', 'QMVZMOYGBJPREO2CFCQ6JAFUXQ======', NULL, '2021-10-30 12:48:33.806974');
INSERT INTO public.git_email (id, user_id, email, token, confirmed_at, created_at) VALUES ('88ad9ea8-61b9-46b3-8d81-abaa03bc81c1', '18a2da08-3fdf-11ec-9356-0242ac130003', 'contributor3@flatfeestack.io', 'MUWBX6P4OSGG47DCN4KDMOKP24======', NULL, '2021-10-30 12:49:24.488047');
INSERT INTO public.git_email (id, user_id, email, token, confirmed_at, created_at) VALUES ('d3bcf76c-35d6-462e-9c2d-846600710b12', '1caeda7a-3fdf-11ec-9356-0242ac130003', 'contributor4@flatfeestack.io', 'MZKRSYENLDMNH6SOCFTFLEVXAY======', NULL, '2021-10-30 12:49:38.526135');

-- Auth
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('sponsor1@flatfeestack.io', 'ACSKRVCANHURIBFN47N42KFO5QXGHYHF7GS6D3B6QRUVZN2VIYHTHNHYLLVGO7UOG55DNKPUPQNVNDQ=', 'PEEM5ZGXUGMUDSUVYDU5V4EGIEYTBEJB', NULL, NULL, 'JFP5V6WB4KCCWEHPHBYZQLHLY7KERCU5', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 12:00:23.557543');
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('sponsor2@flatfeestack.io', 'ACML3UAQAO2BGBT2IKNOF32R5AX4YKQ6GTHTBYINVKRSR24DHQA3VGVXAQOZMJQZ6SSLSLMOT4WMHAQ=', 'Y3XIFCTE7SNNZSIYEYOOTTF43WCRQ4GX', NULL, NULL, 'TAYJBPWWRBYIMSCDQPVP3I6JZLYHIIR2', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 11:58:43.222224');
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('sponsor3@flatfeestack.io', 'AAAR63SNBRAYTOIWTRB3PZNYM52M54H43UH4HTNALNEO7XWCYZGKQFWBWMMJILX6IJC6TZ5BUYPP2VQ=', 'MGRKUZF6WEO4IHEBQISJY3PD4MURKIRY', NULL, NULL, 'JPADP2DDLUFB6TG7ULUHBIKWJRIQTCZU', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 11:59:26.367827');
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('sponsor4@flatfeestack.io', 'ABLA4IQBKDIFFHF3QXYNBFGZ5QLWHYBJV3XBJIRVABTX4OCWTARVT3MYOK255TVJIOPY2IJQ4YMD5EY=', '6LCJNYC77G55OVX5ICTJCLUUPJXUSJBO', NULL, NULL, 'IWBVSWM7EKFWZHM7F67D7VUN65GLZQSM', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 11:59:59.690664');
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('sponsor5@flatfeestack.io', 'ABLA4IQBKDIFFHF3QXYNBFGZ5QLWHYBJV3XBJIRVABTX4OCWTARVT3MYOK255TVJIOPY2IJQ4YMD5EY=', '6LCJNYC77G55OVX5ICTJCLUUPJXUSJBO', NULL, NULL, 'IWBVSWM7EKFWZHM7F67D7VUN65GLZQSM', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 11:59:59.690664');
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('contributor1@flatfeestack.io', 'ACSKRVCANHURIBFN47N42KFO5QXGHYHF7GS6D3B6QRUVZN2VIYHTHNHYLLVGO7UOG55DNKPUPQNVNDQ=', 'PEEM5ZGXUGMUDSUVYDU5V4EGIEYTBEJB', NULL, NULL, 'JFP5V6WB4KCCWEHPHBYZQLHLY7KERCU5', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 12:00:23.557543');
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('contributor2@flatfeestack.io', 'ACML3UAQAO2BGBT2IKNOF32R5AX4YKQ6GTHTBYINVKRSR24DHQA3VGVXAQOZMJQZ6SSLSLMOT4WMHAQ=', 'Y3XIFCTE7SNNZSIYEYOOTTF43WCRQ4GX', NULL, NULL, 'TAYJBPWWRBYIMSCDQPVP3I6JZLYHIIR2', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 11:58:43.222224');
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('contributor3@flatfeestack.io', 'AAAR63SNBRAYTOIWTRB3PZNYM52M54H43UH4HTNALNEO7XWCYZGKQFWBWMMJILX6IJC6TZ5BUYPP2VQ=', 'MGRKUZF6WEO4IHEBQISJY3PD4MURKIRY', NULL, NULL, 'JPADP2DDLUFB6TG7ULUHBIKWJRIQTCZU', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 11:59:26.367827');
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('contributor4@flatfeestack.io', 'ABLA4IQBKDIFFHF3QXYNBFGZ5QLWHYBJV3XBJIRVABTX4OCWTARVT3MYOK255TVJIOPY2IJQ4YMD5EY=', '6LCJNYC77G55OVX5ICTJCLUUPJXUSJBO', NULL, NULL, 'IWBVSWM7EKFWZHM7F67D7VUN65GLZQSM', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 11:59:59.690664');

-- Payment Cycle
INSERT INTO public.payment_cycle (id, user_id, seats, freq, days_left, created_at) VALUES ('bc175e5f-54c0-42e1-87f7-bba11f2a4927', '7a760bac-5d84-498b-9757-5c913d35c605', 1, 365, 365, '2021-10-30 12:06:48.13377');
INSERT INTO public.payment_cycle (id, user_id, seats, freq, days_left, created_at) VALUES ('a457216c-8da3-4b22-8f4d-0fcc8b58328a', '67d62066-b966-425d-8060-2a58a17b48c9', 1, 365, 365, '2021-10-30 12:06:48.13377');
INSERT INTO public.payment_cycle (id, user_id, seats, freq, days_left, created_at) VALUES ('790f85ab-1fd1-455c-a788-913e9d1b67cb', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', 1, 365, 365, '2021-10-30 12:08:18.145275');
INSERT INTO public.payment_cycle (id, user_id, seats, freq, days_left, created_at) VALUES ('0e26bf43-b72f-4ac9-9a64-64ec71825aca', 'd994260b-125e-441a-926a-bd0498bfb902', 1, 365, 365, '2021-10-30 12:15:16.547102');
INSERT INTO public.payment_cycle (id, user_id, seats, freq, days_left, created_at) VALUES ('0fb5a829-4b4a-4853-b742-679977bd2314', 'fdc85b12-3fdf-11ec-9356-0242ac130003', 1, 365, 30, '2021-10-30 12:16:16.194388');
INSERT INTO public.payment_cycle (id, user_id, seats, freq, days_left, created_at) VALUES ('e9db8eba-dfd2-4599-b300-365b19812261', 'fdc85b12-3fdf-11ec-9356-0242ac130003', 1, 365, 395, '2021-10-30 12:22:08.291194');

-- Invoice
INSERT INTO public.invoice (id, nowpayments_invoice_id, payment_cycle_id, payment_id, price_amount, price_currency, pay_amount, pay_currency, actually_paid, outcome_amount, outcome_currency, payment_status, freq, invoice_url, created_at, last_update) VALUES ('61c43566-86ed-472c-b211-7a1eb7018bf4', 1, 'a457216c-8da3-4b22-8f4d-0fcc8b58328a', 4734113644, 125470000, 'USD', 25983643, 'ETH', 25983643, 25983643, 'ETH', 'FINISHED', 365, 'https://nowpayments/invoice/1', '2021-10-30 12:06:48.236', '2021-10-30 12:08:12');
INSERT INTO public.invoice (id, nowpayments_invoice_id, payment_cycle_id, payment_id, price_amount, price_currency, pay_amount, pay_currency, actually_paid, outcome_amount, outcome_currency, payment_status, freq, invoice_url, created_at, last_update) VALUES ('bd46c888-6f10-4517-b223-a27f89a43d3d', 2, '790f85ab-1fd1-455c-a788-913e9d1b67cb', 5785558822, 125470000, 'USD', 18808777429, 'XTZ', 18808777429, 18808777429, 'XTZ', 'FINISHED', 365, 'https://nowpayments/invoice/2', '2021-10-30 12:08:18.139', '2021-10-30 12:09:12');
INSERT INTO public.invoice (id, nowpayments_invoice_id, payment_cycle_id, payment_id, price_amount, price_currency, pay_amount, pay_currency, actually_paid, outcome_amount, outcome_currency, payment_status, freq, invoice_url, created_at, last_update) VALUES ('52f431f2-6b0f-4d2c-a54d-3b9875e7ed03', 3, '0e26bf43-b72f-4ac9-9a64-64ec71825aca', 5127184013, 125470000, 'USD', 2642589738, 'NEO', 2642589738, 2642589738, 'NEO', 'FINISHED', 365, 'https://nowpayments/invoice/3', '2021-10-30 12:15:16.662', '2021-10-30 12:16:08');
INSERT INTO public.invoice (id, nowpayments_invoice_id, payment_cycle_id, payment_id, price_amount, price_currency, pay_amount, pay_currency, actually_paid, outcome_amount, outcome_currency, payment_status, freq, invoice_url, created_at, last_update) VALUES ('848d0be7-460b-467b-aaf5-40d220bc57b0', 4, '0fb5a829-4b4a-4853-b742-679977bd2314', 6134730376, 125470000, 'USD', 2135642, 'ETH', 2135642, 2135642, 'ETH', 'FINISHED', 365, 'https://nowpayments/invoice/4', '2021-10-30 12:16:16.184', '2021-10-30 12:17:37');
INSERT INTO public.invoice (id, nowpayments_invoice_id, payment_cycle_id, payment_id, price_amount, price_currency, pay_amount, pay_currency, actually_paid, outcome_amount, outcome_currency, payment_status, freq, invoice_url, created_at, last_update) VALUES ('7b626ccc-7f0c-45e3-911d-549eb5bb94c9', 5, 'e9db8eba-dfd2-4599-b300-365b19812261', 6284361591, 125470000, 'USD', 2642589738, 'NEO', 2642589738, 2642589738, 'NEO', 'FINISHED', 365, 'https://nowpayments/invoice/5', '2021-10-30 12:22:08.422', '2021-10-30 12:23:02');

-- add payment cycle to sponsor
update users set payment_cycle_id = 'bc175e5f-54c0-42e1-87f7-bba11f2a4927' where id = '7a760bac-5d84-498b-9757-5c913d35c605';
update users set payment_cycle_id = 'a457216c-8da3-4b22-8f4d-0fcc8b58328a' where id = '67d62066-b966-425d-8060-2a58a17b48c9';
update users set payment_cycle_id = '790f85ab-1fd1-455c-a788-913e9d1b67cb' where id = '23fbded4-6d6b-446e-8b33-9cefcacc9a01';
update users set payment_cycle_id = '0e26bf43-b72f-4ac9-9a64-64ec71825aca' where id = 'd994260b-125e-441a-926a-bd0498bfb902';
update users set payment_cycle_id = 'e9db8eba-dfd2-4599-b300-365b19812261' where id = 'fdc85b12-3fdf-11ec-9356-0242ac130003';

-- add user balance
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('ce810529-da3d-4910-8a98-647447c1e67b', 'bc175e5f-54c0-42e1-87f7-bba11f2a4927', '7a760bac-5d84-498b-9757-5c913d35c605', NULL, 125470000, 'PAYMENT', 'USD', '1970-01-01', '2021-10-30 12:08:12.091596');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('50349716-4dc4-11ec-81d3-0242ac130003', 'bc175e5f-54c0-42e1-87f7-bba11f2a4927', '7a760bac-5d84-498b-9757-5c913d35c605', NULL, -5020000, 'FEE', 'USD', '1970-01-01', '2021-10-30 12:08:12.091596');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('f3c9f27e-278e-4487-a512-7e0c8e8b45bf', 'a457216c-8da3-4b22-8f4d-0fcc8b58328a', '67d62066-b966-425d-8060-2a58a17b48c9', NULL, 25983643, 'PAYMENT', 'ETH', '1970-01-01', '2021-10-30 12:08:12.091596');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('690b32a8-3fe2-11ec-9356-0242ac130003', '790f85ab-1fd1-455c-a788-913e9d1b67cb', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', NULL, 18808777429, 'PAYMENT', 'XTZ', '1970-01-01', '2021-10-30 12:08:12.091596');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('690b3730-3fe2-11ec-9356-0242ac130003', '0e26bf43-b72f-4ac9-9a64-64ec71825aca', 'd994260b-125e-441a-926a-bd0498bfb902', NULL, 2642589738, 'PAYMENT', 'NEO', '1970-01-01', '2021-10-30 12:08:12.091596');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('0d8af59a-18cc-4eb7-af7f-b3702619d990', '0fb5a829-4b4a-4853-b742-679977bd2314', 'fdc85b12-3fdf-11ec-9356-0242ac130003', NULL, 2135642, 'PAYMENT', 'ETH', '1970-01-01', '2021-10-30 12:08:12.091596');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('690b382a-3fe2-11ec-9356-0242ac130003', '0fb5a829-4b4a-4853-b742-679977bd2314', 'fdc85b12-3fdf-11ec-9356-0242ac130003', NULL, -2135642, 'CLOSE_CYCLE', 'ETH', '1970-01-01', '2021-10-30 12:08:12.091596');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('13c5edde-ab19-4178-b753-651e419281b9', 'e9db8eba-dfd2-4599-b300-365b19812261', 'fdc85b12-3fdf-11ec-9356-0242ac130003', NULL, 2135642, 'CARRY_OVER', 'ETH', '1970-01-01', '2021-10-30 12:08:12.091596');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('690b38f2-3fe2-11ec-9356-0242ac130003', 'e9db8eba-dfd2-4599-b300-365b19812261', 'fdc85b12-3fdf-11ec-9356-0242ac130003', NULL, 2642589738, 'PAYMENT', 'NEO', '1970-01-01', '2021-10-30 12:08:12.091596');

-- insert daily_payment
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('dca81672-1987-406a-8781-5b36b0844603', 'bc175e5f-54c0-42e1-87f7-bba11f2a4927', 'USD', 330000, 365, '2021-10-30 14:08:12.099251');
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('e3aaea5d-ed9c-4d42-9132-c09e492b7b7a', 'a457216c-8da3-4b22-8f4d-0fcc8b58328a', 'ETH', 71188, 365, '2021-10-30 14:08:12.099251');
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('cd4e512a-a84e-48f1-89b6-520b255aeff4', '790f85ab-1fd1-455c-a788-913e9d1b67cb', 'XTZ', 51530897, 365, '2021-10-30 14:09:12.935132');
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('4105e1e9-4dce-4af8-ac57-9a1b75a365be', '0e26bf43-b72f-4ac9-9a64-64ec71825aca', 'NEO', 7239972, 365, '2021-10-30 14:16:08.703374');
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('bee34584-4dd7-47d3-95a2-603667f96ace', '0fb5a829-4b4a-4853-b742-679977bd2314', 'ETH', 71188, 30, '2021-10-30 14:17:37.707852');
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('2115e768-6146-4926-befe-36daac097bca', 'e9db8eba-dfd2-4599-b300-365b19812261', 'ETH', 71188, 30, '2021-10-30 14:23:02.682054');
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('7fdbfeea-ac4d-4b56-8725-4dfa0383cddf', 'e9db8eba-dfd2-4599-b300-365b19812261', 'NEO', 7239972, 365, '2021-10-30 14:17:37.713275');

-- wallet_address
INSERT INTO public.wallet_address (id, user_id, currency, address, is_deleted) VALUES ('acab0469-b792-4226-9cbf-d08ddaa77bee', 'd263a9f0-3fde-11ec-9356-0242ac130003', 'ETH', 'contributor_1_eth_address', false);
INSERT INTO public.wallet_address (id, user_id, currency, address, is_deleted) VALUES ('5069fbf4-8739-4e5f-bed7-d2dcdd943820', '14b8f83c-3fdf-11ec-9356-0242ac130003', 'ETH', 'contributor_2_eth_address', false);
INSERT INTO public.wallet_address (id, user_id, currency, address, is_deleted) VALUES ('690b39a6-3fe2-11ec-9356-0242ac130003', '14b8f83c-3fdf-11ec-9356-0242ac130003', 'XTZ', 'contributor_2_xtz_address', false);
INSERT INTO public.wallet_address (id, user_id, currency, address, is_deleted) VALUES ('690b3a5a-3fe2-11ec-9356-0242ac130003', '14b8f83c-3fdf-11ec-9356-0242ac130003', 'NEO', 'contributor_2_neo_address', false);
INSERT INTO public.wallet_address (id, user_id, currency, address, is_deleted) VALUES ('11ddc5d2-a2b2-4062-ab32-8688ef62a06d', '18a2da08-3fdf-11ec-9356-0242ac130003', 'ETH', 'contributor_3_eth_address', false);
INSERT INTO public.wallet_address (id, user_id, currency, address, is_deleted) VALUES ('690b3b0e-3fe2-11ec-9356-0242ac130003', '18a2da08-3fdf-11ec-9356-0242ac130003', 'XTZ', 'contributor_3_xtz_address', false);
INSERT INTO public.wallet_address (id, user_id, currency, address, is_deleted) VALUES ('690b3f96-3fe2-11ec-9356-0242ac130003', '18a2da08-3fdf-11ec-9356-0242ac130003', 'NEO', 'contributor_3_neo_address', false);
INSERT INTO public.wallet_address (id, user_id, currency, address, is_deleted) VALUES ('ea1457b6-daff-4077-a154-15130b7f15d0', '1caeda7a-3fdf-11ec-9356-0242ac130003', 'ETH', 'contributor_4_eth_address', false);
INSERT INTO public.wallet_address (id, user_id, currency, address, is_deleted) VALUES ('690b3eec-3fe2-11ec-9356-0242ac130003', '1caeda7a-3fdf-11ec-9356-0242ac130003', 'XTZ', 'contributor_4_xtz_address', false);
INSERT INTO public.wallet_address (id, user_id, currency, address, is_deleted) VALUES ('690b3e38-3fe2-11ec-9356-0242ac130003', '1caeda7a-3fdf-11ec-9356-0242ac130003', 'NEO', 'contributor_4_neo_address', false);

-- repos
INSERT INTO public.repo (id, orig_id, url, git_url, branch, name, description, tags, score, source, created_at) VALUES ('bc4c1eb3-3e6c-4d4a-baaa-0cc521a46730', 14244709, 'https://github.com/repo1', 'https://github.com/repo1', 'main', 'flatfeestack/repo1', 'Flattfeestack Test Repo1', '\x0eff81040102ff8200010c010c000004ff820000', 1, 'github', '2021-10-30 12:27:00.652905');
INSERT INTO public.repo (id, orig_id, url, git_url, branch, name, description, tags, score, source, created_at) VALUES ('b0f2d92c-3891-41ad-bca2-48deb5ed0011', 11730342, 'https://github.com/repo2', 'https://github.com/repo2', 'main', 'flatfeestack/repo2', 'Flattfeestack Test Repo2', '\x0eff81040102ff8200010c010c000004ff820000', 1, 'github', '2021-10-30 12:38:30.358777');
INSERT INTO public.repo (id, orig_id, url, git_url, branch, name, description, tags, score, source, created_at) VALUES ('d0a4aeb7-9430-411b-b944-65afdee0b0f4', 41160453, 'https://github.com/repo3', 'https://github.com/repo3', 'main', 'flatfeestack/repo3', 'Flattfeestack Test Repo3', '\x0eff81040102ff8200010c010c000004ff820000', 1, 'github', '2021-10-30 12:38:52.050538');
INSERT INTO public.repo (id, orig_id, url, git_url, branch, name, description, tags, score, source, created_at) VALUES ('d081217c-9833-4372-bdf3-a5fc863c1ed3', 10270250, 'https://github.com/repo4', 'https://github.com/repo4', 'main', 'flatfeestack/repo4', 'Flattfeestack Test Repo4', '\x0eff81040102ff8200010c010c000004ff820000', 1, 'github', '2021-10-30 12:26:12.322133');

-- sponsor_event
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('ee49595f-b661-4459-ba91-005efcbde80b', 'bc4c1eb3-3e6c-4d4a-baaa-0cc521a46730', '7a760bac-5d84-498b-9757-5c913d35c605', '2021-10-01 00:00:00.000000', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('690b40fe-3fe2-11ec-9356-0242ac130003', 'b0f2d92c-3891-41ad-bca2-48deb5ed0011', '7a760bac-5d84-498b-9757-5c913d35c605', '2021-10-01 00:00:00.000000', '9999-01-01 00:00:00');

INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('ad688eb9-24f0-4322-bf76-ece9756659ef', 'bc4c1eb3-3e6c-4d4a-baaa-0cc521a46730', '67d62066-b966-425d-8060-2a58a17b48c9', '2021-10-05 00:00:00.000000', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('2dbab59c-5d89-4e9a-9a1e-0f5044aa6f29', 'b0f2d92c-3891-41ad-bca2-48deb5ed0011', '67d62066-b966-425d-8060-2a58a17b48c9', '2021-10-05 00:00:00.000000', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('ff669803-b36a-4d3b-a2d8-3d026c0136d5', 'd0a4aeb7-9430-411b-b944-65afdee0b0f4', '67d62066-b966-425d-8060-2a58a17b48c9', '2021-10-05 00:00:00.000000', '9999-01-01 00:00:00');

INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('a4cb7038-1055-4f3d-8d61-3653fd871569', 'd0a4aeb7-9430-411b-b944-65afdee0b0f4', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', '2021-10-10 00:00:00.000000', '9999-01-01 00:00:00');

INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('f54173bb-d683-4cf5-a026-0c49ad95acc3', 'd0a4aeb7-9430-411b-b944-65afdee0b0f4', 'd994260b-125e-441a-926a-bd0498bfb902', '2021-10-18 00:00:00.000000', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('5c1cf33c-feca-4c17-9699-056c626de97b', 'd081217c-9833-4372-bdf3-a5fc863c1ed3', 'd994260b-125e-441a-926a-bd0498bfb902', '2021-10-18 00:00:00.000000', '9999-01-01 00:00:00');

INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('775b13c3-4444-4f0b-b523-c6a622324aba', 'b0f2d92c-3891-41ad-bca2-48deb5ed0011', 'fdc85b12-3fdf-11ec-9356-0242ac130003', '2021-10-18 00:00:00.000000', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('f6d4427c-1601-49d4-bb81-57df8dcbc192', 'd0a4aeb7-9430-411b-b944-65afdee0b0f4', 'fdc85b12-3fdf-11ec-9356-0242ac130003', '2021-10-18 00:00:00.000000', '9999-01-01 00:00:00');



-- analysis_request
INSERT INTO public.analysis_request (id, repo_id, date_from, date_to, branch, created_at) VALUES ('df5586a6-9b7e-481a-bc5b-583b09a39084', 'bc4c1eb3-3e6c-4d4a-baaa-0cc521a46730', '2021-10-01 00:00:00.000000', '2021-10-01 00:00:00.000000', 'master', '2021-10-01 00:00:00.000000');
INSERT INTO public.analysis_request (id, repo_id, date_from, date_to, branch, created_at) VALUES ('d05b5925-3432-4233-baef-325f301c520c', 'b0f2d92c-3891-41ad-bca2-48deb5ed0011', '2021-10-01 00:00:00.000000', '2021-10-01 00:00:00.000000', 'master', '2021-10-01 00:00:00.000000');
INSERT INTO public.analysis_request (id, repo_id, date_from, date_to, branch, created_at) VALUES ('5ee7c5ca-6400-472e-ae55-7c6045bd6936', 'd0a4aeb7-9430-411b-b944-65afdee0b0f4', '2021-10-01 00:00:00.000000', '2021-10-01 00:00:00.000000', 'master', '2021-10-01 00:00:00.000000');
INSERT INTO public.analysis_request (id, repo_id, date_from, date_to, branch, created_at) VALUES ('dba32194-3555-4587-8032-a37712ba0b9d', 'd081217c-9833-4372-bdf3-a5fc863c1ed3', '2021-10-01 00:00:00.000000', '2021-10-01 00:00:00.000000', 'master', '2021-10-01 00:00:00.000000');

-- analysis_response
INSERT INTO public.analysis_response (id, analysis_request_id, git_email, git_name, weight, created_at) VALUES ('b6eb15fe-4476-11ec-81d3-0242ac130003', 'df5586a6-9b7e-481a-bc5b-583b09a39084', 'contributor1@flatfeestack.io', 'Contributor 1', 0.6, '2021-10-01 00:00:00.000000');
INSERT INTO public.analysis_response (id, analysis_request_id, git_email, git_name, weight, created_at) VALUES ('b6eb1824-4476-11ec-81d3-0242ac130003', 'df5586a6-9b7e-481a-bc5b-583b09a39084', 'contributor2@flatfeestack.io', 'Contributor 2', 0.4, '2021-10-01 00:00:00.000000');

INSERT INTO public.analysis_response (id, analysis_request_id, git_email, git_name, weight, created_at) VALUES ('b6eb1914-4476-11ec-81d3-0242ac130003', 'd05b5925-3432-4233-baef-325f301c520c', 'contributor1@flatfeestack.io', 'Contributor 1', 0.4, '2021-10-01 00:00:00.000000');
INSERT INTO public.analysis_response (id, analysis_request_id, git_email, git_name, weight, created_at) VALUES ('b6eb19dc-4476-11ec-81d3-0242ac130003', 'd05b5925-3432-4233-baef-325f301c520c', 'contributor2@flatfeestack.io', 'Contributor 2', 0.6, '2021-10-01 00:00:00.000000');

INSERT INTO public.analysis_response (id, analysis_request_id, git_email, git_name, weight, created_at) VALUES ('b6eb1aa4-4476-11ec-81d3-0242ac130003', '5ee7c5ca-6400-472e-ae55-7c6045bd6936', 'contributor2@flatfeestack.io', 'Contributor 2', 0.2, '2021-10-01 00:00:00.000000');
INSERT INTO public.analysis_response (id, analysis_request_id, git_email, git_name, weight, created_at) VALUES ('b6eb1b62-4476-11ec-81d3-0242ac130003', '5ee7c5ca-6400-472e-ae55-7c6045bd6936', 'contributor3@flatfeestack.io', 'Contributor 3', 0.5, '2021-10-01 00:00:00.000000');
INSERT INTO public.analysis_response (id, analysis_request_id, git_email, git_name, weight, created_at) VALUES ('b6eb1c16-4476-11ec-81d3-0242ac130003', '5ee7c5ca-6400-472e-ae55-7c6045bd6936', 'contributor4@flatfeestack.io', 'Contributor 4', 0.3, '2021-10-01 00:00:00.000000');

INSERT INTO public.analysis_response (id, analysis_request_id, git_email, git_name, weight, created_at) VALUES ('b6eb1ee6-4476-11ec-81d3-0242ac130003', 'dba32194-3555-4587-8032-a37712ba0b9d', 'not@flatfeestack.io', 'Not at flatfeestack', 0.3, '2021-10-01 00:00:00.000000');

select tmp.contributor_name, sum(tmp.balance) from (
                                                       SELECT
                                                           s.user_id,
                                                           s.repo_id,
                                                           res.git_email as contributor_email,
                                                           res.git_name as contributor_name,
                                                           res.weight as contributor_weight,
                                                           g.user_id as contributor_user_id,
                                                           drb.currency,
                                                           CASE WHEN g.user_id IS NULL THEN
                                                                    FLOOR(drb.balance * res.weight / (drw.weight + res.weight))
                                                                ELSE
                                                                    FLOOR(drb.balance * res.weight / drw.weight)
                                                               END as balance,
                                                           drb.day as day,
    now() as created_at
                                                       FROM sponsor_event AS s
                                                           JOIN (
                                                           SELECT id, MAX(date_to) as date_to, ARRAY_AGG(date_from) as dates_from, repo_id
                                                           FROM analysis_request
                                                           WHERE date_to <= '2021-10-01T23:59:59'
                                                           GROUP BY repo_id, id
                                                           ) AS tmp on tmp.repo_id = s.repo_id
                                                           JOIN analysis_response as res on res.analysis_request_id = tmp.id
                                                           JOIN daily_repo_balance as drb on drb.repo_id = s.repo_id
                                                           JOIN daily_repo_weight as drw on drw.repo_id = s.repo_id
                                                           LEFT JOIN git_email g ON g.email = res.git_email
                                                       WHERE drb.day <= '2021-09-30' and drw.day <= '2021-09-30' and drw.day = drb.day
                                                       order by drb.day, s.repo_id
                                                   ) as tmp group by tmp.contributor_name
-- s.user_id, s.repo_id, drb.day, drb.currency, res.git_email, res.weight, drb.balance * res.weight
-- select repo_id, currency, day, sum(balance) OVER (PARTITION BY currency order by day) as cum_amt from daily_repo_balance
-- where day <= '2021-10-01'
-- order by day

select * from analysis_response