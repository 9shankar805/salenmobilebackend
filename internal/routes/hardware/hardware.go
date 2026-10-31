package hardware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"smg-backend/internal/middleware"
	"smg-backend/internal/models"
)

type HardwareRouter struct {
	DB *gorm.DB
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	h := &HardwareRouter{DB: db}
	h.seedHardwareDefaults()

	// Hardware APIs — require MOBILE plan (or VIP/ANNUAL/LIFETIME)
	hw := r.Group("/hardware", middleware.RequireSubscription(db, "MOBILE"))
	{
		hw.GET("/models", h.GetModels)
		hw.GET("/diode-values", h.GetDiodeValues)
		hw.GET("/jumpers", h.GetJumpers)
	}

	// Admin Hardware Management APIs
	admHw := r.Group("/api/admin/hardware", middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		admHw.GET("/models", h.GetModels)
		admHw.POST("/models", h.CreateModel)
		admHw.DELETE("/models/:id", h.DeleteModel)

		admHw.GET("/diode-values", h.GetDiodeValues)
		admHw.POST("/diode-values", h.CreateDiodeValue)
		admHw.DELETE("/diode-values/:id", h.DeleteDiodeValue)

		admHw.GET("/jumpers", h.GetJumpers)
		admHw.POST("/jumpers", h.CreateJumper)
		admHw.DELETE("/jumpers/:id", h.DeleteJumper)
	}
}

func (h *HardwareRouter) GetModels(c *gin.Context) {
	var modelsList []models.HardwareModel
	brand := c.Query("brand")

	query := h.DB.Order("name ASC")
	if brand != "" {
		query = query.Where("LOWER(brand) = LOWER(?)", brand)
	}

	if err := query.Find(&modelsList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": modelsList})
}

func (h *HardwareRouter) GetDiodeValues(c *gin.Context) {
	modelID := c.Query("modelId")
	connector := c.Query("connector")

	var values []models.DiodeValue
	query := h.DB.Order("pin_number ASC")

	if modelID != "" {
		query = query.Where("model_id = ?", modelID)
	}
	if connector != "" {
		query = query.Where("connector_name = ?", connector)
	}

	if err := query.Find(&values).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": values})
}

func (h *HardwareRouter) GetJumpers(c *gin.Context) {
	modelID := c.Query("modelId")
	category := c.Query("category")

	var jumpers []models.JumperTrace
	query := h.DB.Order("title ASC")

	if modelID != "" {
		query = query.Where("model_id = ?", modelID)
	}
	if category != "" {
		query = query.Where("LOWER(category) = LOWER(?)", category)
	}

	if err := query.Find(&jumpers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": jumpers})
}

func (h *HardwareRouter) CreateModel(c *gin.Context) {
	var item models.HardwareModel
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item.ID == "" {
		item.ID = "hw_" + uuid.New().String()[:8]
	}
	item.CreatedAt = time.Now()

	if err := h.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

func (h *HardwareRouter) CreateDiodeValue(c *gin.Context) {
	var item models.DiodeValue
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item.ID == "" {
		item.ID = "dv_" + uuid.New().String()[:8]
	}

	if err := h.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

func (h *HardwareRouter) CreateJumper(c *gin.Context) {
	var item models.JumperTrace
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item.ID == "" {
		item.ID = "jp_" + uuid.New().String()[:8]
	}

	if err := h.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

func (h *HardwareRouter) DeleteModel(c *gin.Context) {
	id := c.Param("id")
	h.DB.Where("model_id = ?", id).Delete(&models.DiodeValue{})
	h.DB.Where("model_id = ?", id).Delete(&models.JumperTrace{})
	h.DB.Delete(&models.HardwareModel{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Model and related data deleted"})
}

func (h *HardwareRouter) DeleteDiodeValue(c *gin.Context) {
	id := c.Param("id")
	h.DB.Delete(&models.DiodeValue{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Diode value deleted"})
}

func (h *HardwareRouter) DeleteJumper(c *gin.Context) {
	id := c.Param("id")
	h.DB.Delete(&models.JumperTrace{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Jumper trace deleted"})
}

func (h *HardwareRouter) seedHardwareDefaults() {
	var count int64
	h.DB.Model(&models.HardwareModel{}).Count(&count)
	if count > 0 {
		return
	}

	defaultModels := []models.HardwareModel{
		{
			ID:                "rn11",
			Name:              "Redmi Note 11 (Spes)",
			Brand:             "Xiaomi",
			Processor:         "Qualcomm Snapdragon 680 (6nm)",
			PMIC:              "PM6150 / PM6150L",
			BoardArchitecture: "Single Layer PCB with Dual Shield Cans",
			Description:       "Widely repaired mid-range device with known PMIC charging dropouts and audio IC solder crack issues.",
			CommonFault:       "Dead condition with 0.05A to 0.12A power supply hanging pulse (PM6150 VREG_S4 rail short).",
			CreatedAt:         time.Now(),
		},
		{
			ID:                "ip13p",
			Name:              "iPhone 13 Pro (A2638)",
			Brand:             "Apple",
			Processor:         "Apple A15 Bionic (5nm)",
			PMIC:              "Apple Custom Dialog PMU",
			BoardArchitecture: "Sandwich Dual-Layer Interposer Logic Board",
			Description:       "High-end dual-board device requiring hot air separating pre-heater and interposer reball stencil.",
			CommonFault:       "White screen of death after iOS update or drop impact (Display 120Hz ProMotion clock trace severed).",
			CreatedAt:         time.Now(),
		},
		{
			ID:                "a52",
			Name:              "Samsung Galaxy A52 (SM-A525F)",
			Brand:             "Samsung",
			Processor:         "Snapdragon 720G",
			PMIC:              "PM7150 / PM7150A",
			BoardArchitecture: "Single PCB with Sub-Board Interconnect FPC",
			Description:       "Common charging flex thermal disconnect issue, slow charging warning, and microphone dropouts.",
			CommonFault:       "No charging / triangle yellow temperature warning due to 100K NTC thermistor rail open circuit.",
			CreatedAt:         time.Now(),
		},
	}
	h.DB.Create(&defaultModels)

	// Seed Diode Values for rn11
	diodeValues := []models.DiodeValue{
		{ID: "dv1", ModelID: "rn11", ConnectorName: "Display FPC (30-Pin)", PinNumber: 1, PinName: "GND", LineType: "GROUND", NormalReadingMv: -1, TestVoltage: "0.00V", FunctionDescription: "Ground Reference Pad", FaultDiagnosis: "If reading > 0mV, ground solder pad is cracked."},
		{ID: "dv2", ModelID: "rn11", ConnectorName: "Display FPC (30-Pin)", PinNumber: 2, PinName: "DSI_CLK_N", LineType: "CLOCK", NormalReadingMv: 435, TestVoltage: "1.20V", FunctionDescription: "MIPI Display Serial Clock Negative", FaultDiagnosis: "If OL: Black screen with backlight on. Check EMI filter."},
		{ID: "dv3", ModelID: "rn11", ConnectorName: "Display FPC (30-Pin)", PinNumber: 3, PinName: "DSI_CLK_P", LineType: "CLOCK", NormalReadingMv: 432, TestVoltage: "1.20V", FunctionDescription: "MIPI Display Serial Clock Positive", FaultDiagnosis: "If short to GND: CPU MIPI transmitter failure."},
		{ID: "dv4", ModelID: "rn11", ConnectorName: "Display FPC (30-Pin)", PinNumber: 8, PinName: "VREG_L11_1P8", LineType: "POWER", NormalReadingMv: 510, TestVoltage: "1.80V", FunctionDescription: "Display Logic Power Rail", FaultDiagnosis: "If 0V: Check L11 output capacitor near PM6150."},
		{ID: "dv5", ModelID: "rn11", ConnectorName: "Display FPC (30-Pin)", PinNumber: 15, PinName: "LCM_AVDD_5P5", LineType: "POWER", NormalReadingMv: 620, TestVoltage: "5.50V", FunctionDescription: "Display Driver Positive Boost Rail", FaultDiagnosis: "If < 200mV: Boost diode leak. No display graphics."},
	}
	h.DB.Create(&diodeValues)

	// Seed Jumper Traces for rn11
	jumperTraces := []models.JumperTrace{
		{
			ID:           "jp1",
			ModelID:      "rn11",
			Title:        "LCD Backlight Anode LED+ Jumper Trace",
			Category:     "DISPLAY",
			Difficulty:   "Level 2 (Micro-Soldering)",
			Symptom:      "Phone powers on, vibrates, makes sound, but display has zero light.",
			RootCause:    "Burned boost coil L2402 or cracked FPC Pin 1 anode copper trace after water drop.",
			StartPoint:   "Diode D2401 Cathode Pad",
			EndPoint:     "FPC Pin 1 (LED+ Anode Pad)",
			WireThickness: "0.02mm Silver Enamelled Jumper Wire",
			IronTemp:     "350°C",
			UVMask:       "Apply green UV solder mask cure under 365nm UV lamp for 60 seconds.",
			StepByStep:   `["Remove motherboard shield near backlight boost section","Locate diode D2401 near power IC","Scrape solder mask on cathode pad","Solder 0.02mm wire to diode cathode","Route cleanly along ground copper pour","Solder other end directly to Pin 1 of LCD FPC connector","Apply UV curing mask over wire"]`,
			StartX:       0.28,
			StartY:       0.34,
			EndX:         0.78,
			EndY:         0.82,
			WireColorHex: "#EF4444",
		},
	}
	h.DB.Create(&jumperTraces)
}
