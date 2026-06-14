package data

type Paciente struct {
	ID         int      `json:"id"`
	Nombre     string   `json:"nombre"`
	TipoSangre string   `json:"tipo_sangre"`
	Prioridad  string   `json:"prioridad"`
	HospitalID int      `json:"hospital_id"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
	DeletedAt  *string  `json:"deleted_at,omitempty"`
}

func (p *Paciente) GetID() int              { return p.ID }
func (p *Paciente) SetID(id int)            { p.ID = id }
func (p *Paciente) SetCreatedAt(t string)   { p.CreatedAt = t }
func (p *Paciente) SetUpdatedAt(t string)   { p.UpdatedAt = t }
func (p *Paciente) IsActive() bool          { return p.DeletedAt == nil }
func (p *Paciente) SetDeleted(t string)     { p.DeletedAt = &t }

type Donante struct {
	ID         int      `json:"id"`
	Nombre     string   `json:"nombre"`
	TipoSangre string   `json:"tipo_sangre"`
	Edad       int      `json:"edad"`
	Detalle    string   `json:"detalle"`
	HospitalID int      `json:"hospital_id"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
	DeletedAt  *string  `json:"deleted_at,omitempty"`
}

func (d *Donante) GetID() int              { return d.ID }
func (d *Donante) SetID(id int)            { d.ID = id }
func (d *Donante) SetCreatedAt(t string)   { d.CreatedAt = t }
func (d *Donante) SetUpdatedAt(t string)   { d.UpdatedAt = t }
func (d *Donante) IsActive() bool          { return d.DeletedAt == nil }
func (d *Donante) SetDeleted(t string)     { d.DeletedAt = &t }

type Organo struct {
	ID             int      `json:"id"`
	Tipo           string   `json:"tipo"`
	Compatibilidad string   `json:"compatibilidad"`
	Estado         string   `json:"estado"`
	DonanteID      int      `json:"donante_id"`
	HospitalID     int      `json:"hospital_id"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	DeletedAt      *string  `json:"deleted_at,omitempty"`
}

func (o *Organo) GetID() int              { return o.ID }
func (o *Organo) SetID(id int)            { o.ID = id }
func (o *Organo) SetCreatedAt(t string)   { o.CreatedAt = t }
func (o *Organo) SetUpdatedAt(t string)   { o.UpdatedAt = t }
func (o *Organo) IsActive() bool          { return o.DeletedAt == nil }
func (o *Organo) SetDeleted(t string)     { o.DeletedAt = &t }

type Trasplante struct {
	ID         int      `json:"id"`
	PacienteID int      `json:"paciente_id"`
	OrganoID   int      `json:"organo_id"`
	Fecha      string   `json:"fecha"`
	Estado     string   `json:"estado"`
	HospitalID int      `json:"hospital_id"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
	DeletedAt  *string  `json:"deleted_at,omitempty"`
}

func (t *Trasplante) GetID() int              { return t.ID }
func (t *Trasplante) SetID(id int)            { t.ID = id }
func (t *Trasplante) SetCreatedAt(s string)   { t.CreatedAt = s }
func (t *Trasplante) SetUpdatedAt(s string)   { t.UpdatedAt = s }
func (t *Trasplante) IsActive() bool          { return t.DeletedAt == nil }
func (t *Trasplante) SetDeleted(s string)     { t.DeletedAt = &s }
