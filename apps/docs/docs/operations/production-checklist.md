# Production Deployment Checklist

Complete checklist for deploying Real Staging AI to production on Render with Backblaze B2.

## Pre-Deployment

### Infrastructure Setup

- [ ] **Render Account**
  - [x] Sign up at [render.com](https://render.com)
  - [ ] Add payment method
  - [x] Verify account

- [ ] **Backblaze B2 Storage**
  - [ ] Create account at [backblaze.com](https://www.backblaze.com/b2/cloud-storage.html)
  - [ ] Add payment method
  - [ ] Create bucket: `realstaging-prod` (Private)
  - [ ] Note bucket region (e.g., `us-west-004`)
  - [ ] Create Application Key with bucket access
  - [ ] Save `keyID` and `applicationKey` securely
  - [ ] Configure CORS for your frontend domain

- [ ] **Auth0 Configuration**
  - [ ] Create production application
  - [ ] Configure allowed callback URLs for production
  - [ ] Configure allowed logout URLs
  - [ ] Configure allowed web origins
  - [ ] Enable refresh token rotation
  - [ ] Save Domain and Audience values

- [ ] **Stripe Configuration**
  - [ ] Complete business verification
  - [ ] Switch to Live Mode
  - [ ] Get live API keys (secret and publishable)
  - [ ] Create webhook endpoint for production URL
  - [ ] Save webhook secret
  - [ ] Configure products and pricing
  - [ ] Test checkout flow in live mode

- [ ] **Replicate Account**
  - [ ] Create account at [replicate.com](https://replicate.com)
  - [ ] Get API token
  - [ ] Add payment method
  - [ ] Verify model access (`qwen/qwen-image-edit`)

### Code Preparation

- [ ] **Update Configuration**
  - [ ] Review `render.yaml` in repository root
  - [ ] Update region to your preferred location
  - [ ] Update S3 endpoint to match B2 bucket region
  - [ ] Update S3 bucket name
  - [ ] Update frontend URL
  - [ ] Commit changes to main branch

- [ ] **Version Control**
  - [ ] Tag release version: `git tag -a v1.0.0 -m "Production release"`
  - [ ] Push tags: `git push origin --tags`
  - [ ] Ensure all changes are committed and pushed

- [ ] **Docker Images**
  - [ ] Test local Docker builds: `make build-api` and `make build-worker`
  - [ ] Verify Dockerfiles are production-ready
  - [ ] Check for hardcoded development values

## Deployment

### Render Setup

- [ ] **Connect Repository**
  - [x] Log into Render dashboard
  - [x] Click "New" → "Blueprint"
  - [x] Connect GitHub repository
  - [x] Select main branch
  - [ ] Render detects `render.yaml`
  - [ ] Click "Apply"

- [ ] **Configure Secrets (API Service)**
  - [ ] `AUTH0_DOMAIN`: `your-tenant.us.auth0.com`
  - [ ] `AUTH0_AUDIENCE`: `https://api.yourdomain.com`
  - [ ] `S3_ACCESS_KEY`: B2 keyID
  - [ ] `S3_SECRET_KEY`: B2 applicationKey
  - [ ] `STRIPE_SECRET_KEY`: `sk_live_...`
  - [ ] `STRIPE_WEBHOOK_SECRET`: `whsec_...`
  - [ ] `REPLICATE_API_TOKEN`: `r8_...`

- [ ] **Configure Secrets (Worker Service)**
  - [ ] `S3_ACCESS_KEY`: B2 keyID
  - [ ] `S3_SECRET_KEY`: B2 applicationKey
  - [ ] `REPLICATE_API_TOKEN`: `r8_...`

- [ ] **Database Setup**
  - [ ] Wait for PostgreSQL to provision
  - [ ] Note connection string
  - [ ] Open shell to API service
  - [ ] Run migrations: `/app/migrate up`
  - [ ] Verify: `/app/migrate version`

### Verification

- [ ] **Health Checks**
  - [ ] Check API health: `curl https://realstaging-api.onrender.com/health`
  - [ ] Verify database connection
  - [ ] Verify Redis connection
  - [ ] Check service logs for errors

- [ ] **Functional Testing**
  - [ ] Test user signup/login via Auth0
  - [ ] Test S3 presigned upload to B2
  - [ ] Test image creation and job queueing
  - [ ] Test worker processing
  - [ ] Verify staged images uploaded to B2
  - [ ] Test subscription checkout
  - [ ] Test Stripe webhook processing

## Post-Deployment

### Domain Configuration

- [ ] **Custom Domain (Optional)**
  - [ ] Add custom domain in Render dashboard
  - [ ] Configure DNS records
  - [ ] Wait for SSL certificate provisioning
  - [ ] Test HTTPS access

- [ ] **Update External Services**
  - [ ] Update Auth0 callback URLs to production domain
  - [ ] Update Stripe webhook URLs to production domain
  - [ ] Update frontend environment variables
  - [ ] Test OAuth flow end-to-end

### Monitoring & Alerts

- [ ] **Render Monitoring**
  - [ ] Enable metrics in Render dashboard
  - [ ] Set up alerts for service down
  - [ ] Set up alerts for high error rate
  - [ ] Set up alerts for high response time
  - [ ] Configure notification channels (email, Slack)

- [ ] **Log Management**
  - [ ] Review log output in Render dashboard
  - [ ] Configure log retention
  - [ ] Set up log alerts for critical errors

### Security

- [ ] **Secrets Rotation**
  - [ ] Document secret rotation procedures
  - [ ] Set calendar reminder for quarterly rotation
  - [ ] Test rotation process in staging first

- [ ] **Access Control**
  - [ ] Limit Render dashboard access
  - [ ] Limit Auth0 admin access
  - [ ] Limit Stripe admin access
  - [ ] Document who has access to what

- [ ] **Security Headers**
  - [ ] Verify HTTPS is enforced
  - [ ] Check CORS configuration
  - [ ] Review security headers

### Backup & Recovery

- [ ] **Database Backups**
  - [ ] Verify Render automatic backups are enabled
  - [ ] Set backup retention period
  - [ ] Document restore procedures
  - [ ] Test restore process

- [ ] **Disaster Recovery Plan**
  - [ ] Document rollback procedures
  - [ ] Document data recovery procedures
  - [ ] Keep copy of critical credentials offline
  - [ ] Test recovery procedures

### Performance

- [ ] **Baseline Metrics**
  - [ ] Record initial response times
  - [ ] Record initial error rates
  - [ ] Record initial job processing times
  - [ ] Set performance targets

- [ ] **Optimization**
  - [ ] Review database query performance
  - [ ] Check connection pool settings
  - [ ] Monitor worker concurrency
  - [ ] Plan for scaling thresholds

### Documentation

- [ ] **Runbook**
  - [ ] Document common operations
  - [ ] Document troubleshooting steps
  - [ ] Document escalation procedures
  - [ ] Document on-call procedures

- [ ] **Update Documentation**
  - [ ] Update deployment docs with actual values
  - [ ] Document any issues encountered
  - [ ] Document solutions and workarounds
  - [ ] Update architecture diagrams if needed

## Ongoing Maintenance

### Weekly

- [ ] Review error logs
- [ ] Check service health metrics
- [ ] Review billing and costs
- [ ] Monitor disk usage

### Monthly

- [ ] Review security logs
- [ ] Check for service updates
- [ ] Review backup integrity
- [ ] Review cost optimization opportunities

### Quarterly

- [ ] Rotate secrets
- [ ] Review and update documentation
- [ ] Review access control lists
- [ ] Performance review and optimization
- [ ] Capacity planning review

## Rollback Plan

If deployment fails or issues arise:

1. **Immediate Actions**
   - [ ] Stop new deployments
   - [ ] Assess impact and severity
   - [ ] Communicate to stakeholders

2. **Rollback Steps**
   - [ ] Revert to previous Render deployment
   - [ ] Or roll back git commit and redeploy
   - [ ] Roll back database migrations if needed
   - [ ] Verify services are healthy
   - [ ] Test critical user flows

3. **Post-Incident**
   - [ ] Document what went wrong
   - [ ] Identify root cause
   - [ ] Create action items to prevent recurrence
   - [ ] Update runbook with lessons learned

## Cost Monitoring

### Expected Monthly Costs

- **Render Services**: ~$31/month (starter tier)
  - API: $7/month
  - Worker: $7/month
  - PostgreSQL: $7/month
  - Redis: $10/month

- **Backblaze B2**: ~$5-20/month
  - Storage: $0.005/GB
  - Bandwidth: First 3x storage free
  - API calls: Generous free tier

- **Replicate AI**: Variable
  - ~$0.011 per image processed
  - Estimate based on expected volume

- **Total Infrastructure**: ~$40-60/month (before AI costs)

### Cost Alerts

- [ ] Set up billing alerts in Render
- [ ] Set up billing alerts in Backblaze
- [ ] Monitor Replicate usage and costs
- [ ] Review monthly cost reports

## Success Criteria

Deployment is successful when:

- [x] All services are running and healthy
- [x] Health endpoints return 200 OK
- [x] Users can sign up and log in
- [x] Users can upload images
- [x] Images are processed successfully
- [x] Users can subscribe and checkout
- [x] Stripe webhooks are processing correctly
- [x] No critical errors in logs
- [x] Response times are acceptable (<500ms P95)
- [x] Monitoring and alerts are configured
- [x] Backups are enabled and tested
- [x] Documentation is updated

---

## Additional Resources

- [Deployment Guide](deployment.md) - Detailed deployment instructions
- [Configuration Guide](../guides/configuration.md) - Environment variables
- [Monitoring Guide](monitoring.md) - Observability setup
- [Stripe Integration](../guides/stripe-billing.md) - Payment processing
- [Auth0 Setup](../guides/authentication.md) - Authentication configuration

## Support

If you encounter issues:

1. Check the [Troubleshooting Guide](troubleshooting.md)
2. Review service logs in Render dashboard
3. Check [Render Status](https://status.render.com/)
4. Check [Backblaze Status](https://status.backblaze.com/)
5. Review this checklist for missed steps
