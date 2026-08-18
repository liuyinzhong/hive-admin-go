package router

import (
	"hive-admin-go/controllers"
	"hive-admin-go/middleware"

	"github.com/gin-gonic/gin"
)

// registerProductRoutes 注册产品模块路由，包括 SPU、RP、MP、SKU 和 SKU 价格。
func registerProductRoutes(api *gin.RouterGroup, deps *RouterDeps) {
	productSpuController := controllers.NewProductSpuController()
	productRpController := controllers.NewProductRpController()
	productMpController := controllers.NewProductMpController()
	productSkuController := controllers.NewProductSkuController()
	productSkuPriceController := controllers.NewProductSkuPriceController()
	permissionGuard := deps.PermissionGuard

	product := api.Group("/product", middleware.AuthMiddleware(), deps.DataPermissionMiddleware)
	{
		spus := product.Group("/spus")
		{
			spus.GET("", permissionGuard.Require("product:spu:list"), productSpuController.GetProductSpuList)
			spus.GET("/options", productSpuController.GetProductSpuOptions)
			spus.POST("", permissionGuard.Require("product:spu:create"), productSpuController.CreateProductSpu)
			spus.GET("/:spuId", permissionGuard.Require("product:spu:detail"), productSpuController.GetProductSpu)
			spus.PUT("/:spuId", permissionGuard.Require("product:spu:update"), productSpuController.UpdateProductSpu)
			spus.PUT("/:spuId/status", permissionGuard.Require("product:spu:status"), productSpuController.UpdateProductSpuStatus)
		}
		rps := product.Group("/rps")
		{
			rps.GET("", permissionGuard.Require("product:rp:list"), productRpController.GetProductRpList)
			rps.GET("/options", productRpController.GetProductRpOptions)
			rps.POST("", permissionGuard.Require("product:rp:create"), productRpController.CreateProductRp)
			rps.GET("/:rpId", permissionGuard.Require("product:rp:detail"), productRpController.GetProductRp)
			rps.PUT("/:rpId", permissionGuard.Require("product:rp:update"), productRpController.UpdateProductRp)
			rps.PUT("/:rpId/status", permissionGuard.Require("product:rp:status"), productRpController.UpdateProductRpStatus)
		}
		mps := product.Group("/mps")
		{
			mps.GET("", permissionGuard.Require("product:mp:list"), productMpController.GetProductMpList)
			mps.POST("", permissionGuard.Require("product:mp:create"), productMpController.CreateProductMp)
			mps.GET("/:mpId", permissionGuard.Require("product:mp:detail"), productMpController.GetProductMp)
			mps.PUT("/:mpId", permissionGuard.Require("product:mp:update"), productMpController.UpdateProductMp)
			mps.PUT("/:mpId/status", permissionGuard.Require("product:mp:status"), productMpController.UpdateProductMpStatus)
		}
		skus := product.Group("/skus")
		{
			skus.GET("", permissionGuard.Require("product:sku:list"), productSkuController.GetProductSkuList)
			skus.GET("/options", productSkuController.GetProductSkuOptions)
			skus.POST("", permissionGuard.Require("product:sku:create"), productSkuController.CreateProductSku)
			skus.GET("/:skuId/prices", permissionGuard.Require("product:skuPrice:list"), productSkuPriceController.GetProductSkuPriceList)
			skus.POST("/:skuId/prices", permissionGuard.Require("product:skuPrice:create"), productSkuPriceController.CreateProductSkuPrice)
			skus.PUT("/:skuId/prices/:priceId", permissionGuard.Require("product:skuPrice:update"), productSkuPriceController.UpdateProductSkuPrice)
			skus.PUT("/:skuId/prices/:priceId/status", permissionGuard.Require("product:skuPrice:status"), productSkuPriceController.UpdateProductSkuPriceStatus)
			skus.DELETE("/:skuId/prices/:priceId", permissionGuard.Require("product:skuPrice:delete"), productSkuPriceController.DeleteProductSkuPrice)
			skus.GET("/:skuId/prices/:priceId/tiers", permissionGuard.Require("product:skuPriceTier:list"), productSkuPriceController.GetProductSkuPriceTiers)
			skus.PUT("/:skuId/prices/:priceId/tiers", permissionGuard.Require("product:skuPriceTier:save"), productSkuPriceController.SaveProductSkuPriceTiers)
			skus.GET("/:skuId", permissionGuard.Require("product:sku:detail"), productSkuController.GetProductSku)
			skus.PUT("/:skuId", permissionGuard.Require("product:sku:update"), productSkuController.UpdateProductSku)
			skus.PUT("/:skuId/status", permissionGuard.Require("product:sku:status"), productSkuController.UpdateProductSkuStatus)
		}
	}
}
