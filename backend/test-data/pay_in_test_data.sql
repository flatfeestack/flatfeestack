-- Create Sponsors
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, token, role, created_at) VALUES ('7a760bac-5d84-498b-9757-5c913d35c605', NULL, NULL, NULL, NULL, NULL, NULL, 'sponsor1@flatfeestack.io', NULL, NULL, '6IREUFR7ST3UENSNKOBOWMSIXMMXU===', 'USR', '2021-10-30 12:00:53.293937');

-- Auth
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('sponsor1@flatfeestack.io', 'ACSKRVCANHURIBFN47N42KFO5QXGHYHF7GS6D3B6QRUVZN2VIYHTHNHYLLVGO7UOG55DNKPUPQNVNDQ=', 'PEEM5ZGXUGMUDSUVYDU5V4EGIEYTBEJB', NULL, NULL, 'JFP5V6WB4KCCWEHPHBYZQLHLY7KERCU5', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 12:00:23.557543');
