.PHONY: gen-certs up down logs clean vulnerable secure

CERTS_DIR = certs
FLUTTER_CA_DIR = app/assets/ca

gen-certs:
	@echo "==> Generating TLS certificates with mkcert..."
	@mkdir -p $(CERTS_DIR) $(FLUTTER_CA_DIR)
	mkcert -cert-file $(CERTS_DIR)/server.pem -key-file $(CERTS_DIR)/server-key.pem localhost 127.0.0.1 10.0.2.2
	@echo "==> Copying root CA to Flutter assets..."
	cp "$$(mkcert -CAROOT)/rootCA.pem" $(FLUTTER_CA_DIR)/rootCA.pem
	@echo "==> Done! Certificates generated in $(CERTS_DIR)/"
	@echo "==> Root CA copied to $(FLUTTER_CA_DIR)/rootCA.pem"

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v
	rm -rf $(CERTS_DIR)/*.pem

vulnerable:
	@bash scripts/toggle-vuln.sh vulnerable

secure:
	@bash scripts/toggle-vuln.sh secure
