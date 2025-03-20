# Define the list of plugins
PLUGIN_NAMES = signer router secretskeymanager publisher redis reqpreprocessor nopschemavalidator signvalidator

.PHONY: install-plugins
install-plugins:
ifeq ($(strip $(PLUGIN_NAMES)),)
	@echo "PLUGIN_NAMES is empty. No plugins to install."
else
	./scripts/install-plugin-gcs.sh $(PLUGIN_NAMES)
endif

.PHONY: deploy-bpp
deploy-bpp:
	gcloud beta run deploy bpp-adapter \
	  --image=asia-southeast1-docker.pkg.dev/ondc-seller-dev/onix/adapter:latest \
  	  --region=asia-southeast1 \
  	  --platform=managed \
	  --no-allow-unauthenticated \
	  --add-volume-mount=volume=gcs,mount-path=/mnt/gcs \
  	  --set-env-vars=CONFIG_FILE=/mnt/gcs/configs/bpp.yaml \
  	  --add-volume=name=gcs,type=cloud-storage,bucket=ondc-seller-dev-onix,readonly=true 


.PHONY: deploy-bap
deploy-bap:
	gcloud beta run deploy bap-adapter \
	  --image=asia-southeast1-docker.pkg.dev/ondc-seller-dev/onix/adapter:latest \
  	  --region=asia-southeast1 \
  	  --platform=managed \
	  --no-allow-unauthenticated \
	  --add-volume-mount=vw
  	  --set-env-vars=CONFIG_FILE=/mnt/gcs/configs/bap.yaml \
  	  --add-volume=name=gcs,type=cloud-storage,bucket=ondc-seller-dev-onix,readonly=true 

