# Final stage
FROM amazon/aws-cli:2.31.12

# Copy the binary from builder stage
COPY --from=peakcom/s5cmd s5cmd .

